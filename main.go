package main

import (
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

func main() {
	const dir = "./watch"
	events := []fsnotify.Op{fsnotify.Create, fsnotify.Remove, fsnotify.Rename, fsnotify.Write}

	// Create a channel to receive events
	ch := make(chan fsnotify.Event)

	// Create a wait group to wait for the watcher goroutine to finish
	var wg sync.WaitGroup

	// Start a goroutine to listen to subscribed events
	wg.Add(1)
	go startWatcher(&wg, dir, ch, events...)

	// Start a goroutine to handle events
	wg.Add(1)
	go startEventHandler(&wg, ch)

	// Wait for all goroutines to finish
	wg.Wait()
}

func startWatcher(wg *sync.WaitGroup, dir string, ch chan fsnotify.Event, events ...fsnotify.Op) {
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
			log.Printf("Original Event: %s", event)
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
// You have to take care to create a handler for each event type.
func startEventHandler(wg *sync.WaitGroup, ch chan fsnotify.Event) {
	defer wg.Done()
	for {
		event := <-ch
		log.Printf("Filtered Event: %s", event)
	}
}
