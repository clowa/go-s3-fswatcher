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
	config Configuration
	// Define CLI flags
	sourceFlag = flag.String("source", "", "The directory to upload to s3. Example: /path/to/source")
	bucketFlag = flag.String("bucket", "", "The name of the bucket to upload the files to. Example: my-s3-bucket")
	prefixFlag = flag.String("prefix", "", "The directory to upload to s3. Example: my-prefix/")
	regionFlag = flag.String("region", "us-west-2", "The AWS region to use. Example: us-west-2")
)

func main() {
	// Create a wait group to wait for the watcher goroutine to finish
	var wg sync.WaitGroup

	// Initialize the application
	inizialize()

	// Load configuration values
	config.Load()
	if success := config.Validate(); !success {
		log.Fatal("Failed to load configuration")
	}
	log.Print("Starting S3 File Watcher")

	// Since files only have content after a Write event, we don't need to listen to Create events
	events := []fsnotify.Op{fsnotify.Write}

	// Create a channel to receive events
	ch := make(chan fsnotify.Event)

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

func inizialize() {
	// Log to stdout
	log.SetOutput(os.Stdout)
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

	s3Config := basic.BucketBasics{S3Client: client}

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
				go s3Config.UploadLargeFile(ctx, config.bucket_name, objKey, path)
			} else {
				log.Printf("Uploading file %s (%d bytes) at %s to %s", filename, size, path, objKey)
				go s3Config.UploadFile(ctx, config.bucket_name, objKey, path)
			}
		}
	}
}
