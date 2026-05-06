// Package watcher watches a directory for changes and triggers reloads on a debounce.
package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounce = 150 * time.Millisecond

// Watch starts a goroutine that calls onChange whenever a file in dir is created, modified, removed, or renamed.
// Events are debounced over a short window so that editor "save" patterns (write + chmod + ...) trigger a single reload.
func Watch(ctx context.Context, dir string, onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fmt.Println("Watching", dir)

	err = watcher.Add(dir)
	if err != nil {
		watcher.Close()
		return err
	}

	go run(ctx, watcher, onChange)

	return nil
}

func run(ctx context.Context, watcher *fsnotify.Watcher, onChange func()) {
	defer watcher.Close()

	var timer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return

		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}

			if !isRelevantChange(ev) {
				continue
			}

			if timer == nil {
				timer = time.AfterFunc(debounce, onChange)
			} else {
				timer.Reset(debounce)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			slog.ErrorContext(ctx, "Error",
				slog.String("actor", "watcher"),
				slog.Any("error", err),
			)
		}
	}
}

func isRelevantChange(event fsnotify.Event) bool {
	if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}

	ext := strings.ToLower(filepath.Ext(event.Name))

	return ext == ".yml" || ext == ".yaml"
}
