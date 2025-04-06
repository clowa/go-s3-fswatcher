package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	// Define CLI flags
	sourceFlag = flag.String("source", "", "The directory to upload to s3. Example: /path/to/source")
	// bucketFlag = flag.String("bucket", "", "The name of the bucket to upload the files to. Example: my-s3-bucket")
	// prefixFlag = flag.String("prefix", "", "The directory to upload to s3. Example: my-prefix/")
)

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
	// ToDo: add support for Rename and Delete events
	events := []fsnotify.Op{fsnotify.Write}

	// Create a channel to receive events
	ch := make(chan fsnotify.Event)

	// Start a goroutine to listen to subscribed events
	wg.Add(1)
	go func() {
		defer wg.Done()
		startedFilteredWatcher(config.watch_dir, ch, events...)
	}()

	// Yield the processor to allow other gorotines to run and prevent the main goroutine from exiting
	// runtime.Gosched() // ToDo: kind of ugly, should find a better way to do this
	// Wait for all goroutines to finish
	wg.Wait()
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
