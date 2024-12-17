package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	basic "github.com/clowa/go-s3-fswatcher/lib/s3"
	"github.com/fsnotify/fsnotify"
)

var config configuration

type configuration struct {
	watch_dir     string
	bucket_name   string
	bucket_prefix string
}

func main() {
	log.Print("Starting S3 File Watcher")
	loadConfig()

	// Since files only have content after a Write event, we don't need to listen to Create events
	events := []fsnotify.Op{fsnotify.Write}

	// Create a channel to receive events
	ch := make(chan fsnotify.Event)

	// Create a wait group to wait for the watcher goroutine to finish
	var wg sync.WaitGroup

	// Start a goroutine to listen to subscribed events
	go startedFilteredWatcher(&wg, config.watch_dir, ch, events...)

	// Start a goroutine to handle events
	go startEventHandler(&wg, ch)

	// Yield the processor to allow other gorotines to run and prevent the main goroutine from exiting
	runtime.Gosched()
	// Wait for all goroutines to finish
	wg.Wait()
}

// loadConfig loads configuration values from environment variables and run validation checks
func loadConfig() {

	// Load configuration from environment variables
	config.watch_dir = os.Getenv("WATCH_DIR")
	config.bucket_name = os.Getenv("S3_BUCKET_NAME")
	config.bucket_prefix = os.Getenv("S3_BUCKET_PREFIX")

	// Validate configuration values
	validateConfig()
}

// validateConfig encapsulates the validation logic for the configuration values
func validateConfig() {
	// Check for required environment variables are set
	requiredEnvVars := []string{"AWS_DEFAULT_REGION", "WATCH_DIR", "S3_BUCKET_NAME", "S3_BUCKET_PREFIX"}
	missingEnvVar := false
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingEnvVar = true
			log.Fatalf("Environment variable %s is required", envVar)
		}
	}

	if missingEnvVar {
		log.Fatal("Required configuration values are missing")
		os.Exit(2)
	}

	// Validate source directory
	if _, err := os.Stat(config.watch_dir); os.IsNotExist(err) {
		log.Fatalf("Invalid source directory. Please provide a valid directory path. Example: /path/to/source")
	}

	// Validate bucket name
	if config.bucket_name == "" {
		log.Fatalf("Invalid S3 bucket name. Please provide a valid bucket name. Example: my-s3-bucket")
	}

	// Validate prefix
	if config.bucket_prefix == "" {
		log.Fatalf("Invalid S3 prefix. Please provide a valid prefix. Example: my-prefix/")
	}
}

// startedFilteredWatcher starts a watcher on a directory and filters events based on the provided event list.
func startedFilteredWatcher(wg *sync.WaitGroup, dir string, ch chan fsnotify.Event, events ...fsnotify.Op) {
	wg.Add(1)
	defer wg.Done()

	// Create a watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Add the directory to the watcher
	if err := watcher.Add(dir); err != nil {
		panic(err)
	}

	// Listen to subscribed events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			for _, e := range events {
				if event.Has(e) {
					ch <- event
				}
			}
		case err := <-watcher.Errors:
			panic(err)
		}
	}
}

// startEventHandler reacts to subscribed events.
// Take care to handle the subscribed events in a separate goroutine to avoid blocking the watcher.
func startEventHandler(wg *sync.WaitGroup, ch chan fsnotify.Event) {
	wg.Add(1)
	defer wg.Done()

	// AWS Config
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK default config: %v", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)
	basicConfig := basic.BucketBasics{S3Client: client}

	// Handle events
	for {
		event := <-ch
		switch event.Op {
		case fsnotify.Write:
			log.Printf("Write: %s", event.Name)

			// Get infos about the file firing the event
			filename := filepath.Base(event.Name)
			path := event.Name
			if !filepath.IsAbs(event.Name) {
				path, err = filepath.Abs(event.Name)
				if err != nil {
					log.Fatalf("unable to get absolute path: %v", err)
				}
			}

			// Upload to S3 since object has content
			// On already existing S3 objects, we can do a hash check to avoid unnecessary uploads
			// For simplicity, we'll upload the file on every Write event
			objKey := filepath.Join(config.bucket_prefix, filename)
			info, err := os.Stat(event.Name)
			if err != nil {
				log.Fatalf("unable to get file info: %v", err)
			}
			size := info.Size()

			// Check if the file is larger than 50 MiB. If it is, use the multipart upload to avoid loading whole file into memory
			if info.Size() > 50*1024*1024 {
				log.Printf("Uploading large file %s (%d bytes) at %s to %s", filename, size, path, objKey)
				go basicConfig.UploadLargeFile(wg, context.TODO(), config.bucket_name, objKey, path)
			} else {
				log.Printf("Uploading file %s (%d bytes) at %s to %s", filename, size, path, objKey)
				go basicConfig.UploadFile(wg, context.TODO(), config.bucket_name, objKey, path)
			}
		}
	}
}
