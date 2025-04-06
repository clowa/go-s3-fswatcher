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
	}()

	// Start a goroutine to populate the events with additional information
	wg.Add(1)
	go func() {
		defer wg.Done()
		transformEvents(fsnotifyCh, customNotifyCh)
	}()

	// Start a goroutine which logs the custom events
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range customNotifyCh {
			log.Printf("Custom event received: %s, %s, %s", event.Op, event.Path, event.Destionation)
		}
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
