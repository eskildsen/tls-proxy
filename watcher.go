package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"path/filepath"
)

func setupHostnameWatcher(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(fmt.Errorf("failed setting up filesystem watcher: %w", err))
	}
	defer watcher.Close()

	// Watch the file and assume it is never recreated or deleted (by e.g. an editor).
	// The safest bet is to watch the directory and later filter pr. name in the event.
	// That is however also less performant.

	directory := filepath.Dir(filename)
	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(fmt.Errorf("failed setting up path for filesystem watcher: %w", err))
		return
	}

	log.Printf("filesystem watcher initialized for file %s\n", filename)

	filesystemEventListener(watcher, filename)
}

func filesystemEventListener(watcher *fsnotify.Watcher, filename string) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Fatal("filesystem watcher failed with case E01")
				return
			}

			log.Println("event:", event)
			if event.Name == filename {
				log.Printf("modified file: %s, operation: %s\n", event.Name, event.Op)

				err := loadHostnames(filename)
				if err != nil {
					log.Println("failed loading hostnames (triggered by file watcher):", err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("filesystem watcher got error:", err)
		}
	}
}
