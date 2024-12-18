package main

import (
	"context"
	"flag"
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

var (
	config configuration
	// Define CLI flags
	sourceFlag = flag.String("source", "", "The directory to upload to s3. Example: /path/to/source")
	bucketFlag = flag.String("bucket", "", "The name of the bucket to upload the files to. Example: my-s3-bucket")
	prefixFlag = flag.String("prefix", "", "The directory to upload to s3. Example: my-prefix/")
	regionFlag = flag.String("region", "us-west-2", "The AWS region to use. Example: us-west-2")
)

type configuration struct {
	watch_dir     string
	bucket_name   string
	bucket_prefix string
	aws_region    string
}

func main() {
	if !loadConfig() {
		log.Fatal("Failed to load configuration")
	}
	log.Print("Starting S3 File Watcher")

	// Since files only have content after a Write event, we don't need to listen to Create events
	events := []fsnotify.Op{fsnotify.Write}

	// Create a channel to receive events
	ch := make(chan fsnotify.Event)

	// Create a wait group to wait for the watcher goroutine to finish
	var wg sync.WaitGroup

	// Start a goroutine to listen to subscribed events
	wg.Add(1)
	go func() {
		defer wg.Done()
		startedFilteredWatcher(config.watch_dir, ch, events...)
	}()

	// Start a goroutine to handle events
	wg.Add(1)
	go func() {
		defer wg.Done()
		startEventHandler(ch)
	}()

	// Yield the processor to allow other gorotines to run and prevent the main goroutine from exiting
	runtime.Gosched() // kind of ugly, should find a better way to do this
	// Wait for all goroutines to finish
	wg.Wait()
}

// loadConfig loads configuration values from environment variables and run validation checks.
// It returns true if the configuration values are valid, false otherwise.
func loadConfig() bool {
	// Parse CLI flags
	flag.Parse()

	config.aws_region = *regionFlag

	// Load configuration from CLI flags or environment variables
	if *sourceFlag != "" {
		config.watch_dir = *sourceFlag
	} else {
		config.watch_dir = os.Getenv("WATCH_DIR")
	}

	if *bucketFlag != "" {
		config.bucket_name = *bucketFlag
	} else {
		config.bucket_name = os.Getenv("S3_BUCKET_NAME")
	}

	if *prefixFlag != "" {
		config.bucket_prefix = *prefixFlag
	} else {
		config.bucket_prefix = os.Getenv("S3_BUCKET_PREFIX")
	}

	// Validate configuration values
	return validateConfig()
}

// validateConfig encapsulates the validation logic for the configuration values.
// It returns true if the configuration values are valid, false otherwise.
func validateConfig() bool {
	// Validate source directory
	if _, err := os.Stat(config.watch_dir); os.IsNotExist(err) {
		log.Printf("Invalid source directory. Please provide a valid directory path. Example: /path/to/source")
		return false
	}

	// Validate bucket name
	if config.bucket_name == "" {
		log.Printf("Invalid S3 bucket name. Please provide a valid bucket name. Example: my-s3-bucket")
		return false
	}

	// Validate prefix
	if config.bucket_prefix == "" {
		log.Printf("Invalid S3 prefix. Please provide a valid prefix. Example: my-prefix/")
		return false
	}

	return true
}

// startedFilteredWatcher starts a watcher on a directory and filters events based on the provided event list.
func startedFilteredWatcher(dir string, ch chan fsnotify.Event, events ...fsnotify.Op) {
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
func startEventHandler(ch chan fsnotify.Event) {
	const largeFileThreshold = 50 * 1024 * 1024 // 50 MiB

	// Context for S3 upload
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// AWS Config
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK default config: %v", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = config.aws_region
	})

	basicConfig := basic.BucketBasics{S3Client: client}

	// Handle events
	for {
		event := <-ch
		switch event.Op {
		case fsnotify.Write:
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

			// Check if the file is larger than threshold. If it is, use the multipart upload to avoid loading whole file into memory
			if size == 0 {
				log.Printf("Skipping empty file %s", filename)
			} else if size > largeFileThreshold {
				log.Printf("Uploading large file %s (%d bytes) at %s to %s", filename, size, path, objKey)
				go basicConfig.UploadLargeFile(ctx, config.bucket_name, objKey, path)
			} else {
				log.Printf("Uploading file %s (%d bytes) at %s to %s", filename, size, path, objKey)
				go basicConfig.UploadFile(ctx, config.bucket_name, objKey, path)
			}
		}
	}
}
