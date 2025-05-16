// Package server provides HTTP server and configuration handling logic,
// including live config file watching with debounced reload triggers.
package server

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchConfig sets up a file system watcher on the specified path.
// When the file is written to, it triggers the onChange callback after a debounce delay.
// This prevents rapid repeated reloads due to multiple quick write events.
func watchConfig(path string, onChange func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	var debounceTimer *time.Timer
	var debounceMu sync.Mutex

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write != 0 {
					debounceMu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(300*time.Millisecond, onChange)
					debounceMu.Unlock()
				}
			case err := <-watcher.Errors:
				log.Println("watch error:", err)
			}
		}
	}()

	watcher.Add(path)
}
