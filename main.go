package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	basic "github.com/clowa/go-s3-fswatcher/lib/s3"
	"github.com/fsnotify/fsnotify"
)

var (
	// Define CLI flags
	sourceFlag = flag.String("source", "", "The directory to upload to s3. Example: /path/to/source")
	bucketFlag = flag.String("bucket", "", "The name of the bucket to upload the files to. Example: my-s3-bucket")
	prefixFlag = flag.String("prefix", "", "The directory to upload to s3. Example: my-prefix/")
)

type CustomFileEvent struct {
	Op           fsnotify.Op
	Path         string
	Destionation string
}

func main() {
	// Create a wait group to wait for the watcher goroutine to finish
	var wg sync.WaitGroup

	// Initialize the application
	log.SetOutput(os.Stdout)

	// Load configuration values
	config := NewConfiguration()
	config.Load()
	if success := config.Validate(); !success {
		log.Fatal("Failed to load configuration")
	}
	log.Print("Starting S3 File Watcher")

	// Since files only have content after a Write event, we don't need to listen to Create events
	// ToDo: add support for Delete events
	events := []fsnotify.Op{fsnotify.Write, fsnotify.Rename}

	// Create a channel to receive events
	fsnotifyCh := make(chan fsnotify.Event)
	customNotifyCh := make(chan CustomFileEvent)

	// Start a goroutine to listen to subscribed events
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchAndFilterEvents(config.watch_dir, fsnotifyCh, events...)
		close(fsnotifyCh)
	}()

	// Start a goroutine to populate the events with additional information
	wg.Add(1)
	go func() {
		defer wg.Done()
		transformEvents(fsnotifyCh, customNotifyCh)
		close(customNotifyCh)
	}()

	// Start a goroutine which logs the custom events
	wg.Add(1)
	go func() {
		defer wg.Done()
		handleFileEvents(*config, customNotifyCh)
	}()

	// Yield the processor to allow other gorotines to run and prevent the main goroutine from exiting
	// runtime.Gosched() // ToDo: kind of ugly, should find a better way to do this
	// Wait for all goroutines to finish
	wg.Wait()
}

// watchAndFilterEvents starts a watcher on a directory and filters events based on the provided event list.
func watchAndFilterEvents(dir string, ch chan fsnotify.Event, eventFilter ...fsnotify.Op) {
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
			for _, e := range eventFilter {
				if event.Has(e) {
					log.Printf("Sending event: %s, %s", event.Op, event.Name)
					ch <- event
				}
			}
		case err := <-watcher.Errors:
			panic(err)
		}
	}
}

// transformEvents transforms fsnotify events into custom file events and forwards them to the destination channel.
// It handles rename events by waiting for the corresponding write event to the new file.
func transformEvents(s chan fsnotify.Event, d chan CustomFileEvent) {
	for event := range s {
		var ce CustomFileEvent

		if event.Has(fsnotify.Rename) {
			// wait for the write event to the new named file
			renameEvent := event
			writeEvent := <-s

			ce = CustomFileEvent{
				Op:           renameEvent.Op,
				Path:         renameEvent.Name,
				Destionation: writeEvent.Name,
			}
		} else {
			ce = CustomFileEvent{
				Op:           event.Op,
				Path:         event.Name,
				Destionation: event.Name,
			}
		}

		log.Printf("Transforming event: %s, %s, %s", ce.Op, ce.Path, ce.Destionation)

		d <- ce
	}
}

// handleFileEvents reacts to subscribed events.
// Take care to handle the subscribed events in a separate goroutine to avoid blocking the watcher.
func handleFileEvents(config Configuration, ch chan CustomFileEvent) {
	const largeFileThreshold = 50 * 1024 * 1024 // 50 MiB

	// Context for S3 upload
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an S3 client
	client := s3.NewFromConfig(config.aws_config)

	s3Config := basic.BucketBasics{S3Client: client}

	// Handle events
	for {
		event := <-ch
		switch event.Op {
		// ToDo: race condtion - multiple goroutines may attempt to upload the same file at the same time
		case fsnotify.Write:
			// Get infos about the file firing the event
			path := event.Path
			filename := filepath.Base(path)
			if !filepath.IsAbs(path) {
				_, err := filepath.Abs(path)
				if err != nil {
					log.Fatalf("unable to get absolute path: %v", err)
				}
			}

			// Upload to S3 since object has content
			// On already existing S3 objects, we can do a hash check to avoid unnecessary uploads
			// For simplicity, we'll upload the file on every Write event
			objKey := filepath.Join(config.bucket_prefix, filename)
			info, err := os.Stat(path)
			if err != nil {
				log.Fatalf("unable to get file info: %v", err)
			}
			size := info.Size()

			// Check if the file is larger than threshold. If it is, use the multipart upload to avoid loading whole file into memory
			if size == 0 {
				log.Printf("Skipping empty file %s", filename)
			} else if size > largeFileThreshold {
				log.Printf("Uploading large file %s (%d bytes) at %s to %s", filename, size, path, objKey)

				go func() {
					if err := s3Config.UploadLargeFile(ctx, config.bucket_name, objKey, path); err != nil {
						log.Printf("Failed to upload large file %s: %v", objKey, err)
					}
				}()
			} else {
				log.Printf("Uploading file %s (%d bytes) at %s to %s", filename, size, path, objKey)

				go func() {
					if err := s3Config.UploadFile(ctx, config.bucket_name, objKey, path); err != nil {
						log.Printf("Failed to upload file %s: %v", objKey, err)
					}
				}()
			}

		case fsnotify.Rename:
			// Handle rename event
			path := event.Path
			currentObjectKey := filepath.Join(config.bucket_prefix, filepath.Base(path))
			newObjectKey := filepath.Join(config.bucket_prefix, filepath.Base(event.Destionation))

			log.Printf("Renaming object %s to %s", currentObjectKey, newObjectKey)
			go func() {
				// Rename the object in S3
				err := s3Config.RenameObject(ctx, config.bucket_name, currentObjectKey, newObjectKey)
				if err != nil {
					log.Printf("Failed to rename object %s to %s: %v", currentObjectKey, newObjectKey, err)
				}
			}()

		}
	}
}
