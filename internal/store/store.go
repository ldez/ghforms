// Package store holds the current state of the loaded issue forms and
// notifies subscribers whenever it changes.
package store

import (
	"sync"

	"github.com/ldez/githubformpreview/internal/form"
)

type Loader interface {
	Load(dir string) ([]*form.Form, error)
}

// Snapshot is a point-in-time view of the forms directory:
// either a list of valid forms or the error encountered while loading/validating them.
type Snapshot struct {
	Forms []*form.Form
	Err   error

	// Version increments on each reload; clients can use it for SSE.
	Version uint64
}

type Store struct {
	mu      sync.RWMutex
	dir     string
	current Snapshot

	loader Loader

	subsMu sync.Mutex
	subs   map[chan struct{}]struct{}
}

// New creates a Store and performs the initial load.
func New(dir string, loader Loader) *Store {
	s := &Store{
		dir:    dir,
		subs:   make(map[chan struct{}]struct{}),
		loader: loader,
	}

	s.Reload()

	return s
}

// Get returns the current snapshot.
func (s *Store) Get() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.current
}

// Dir returns the watched directory path.
func (s *Store) Dir() string { return s.dir }

// Reload re-reads and re-validates all forms in the directory.
// Any error becomes part of the snapshot.
// It is not returned to the caller because the goal of live-reload is to display errors in the UI.
func (s *Store) Reload() {
	forms, err := s.loader.Load(s.dir)

	s.mu.Lock()

	s.current = Snapshot{
		Forms:   forms,
		Err:     err,
		Version: s.current.Version + 1,
	}

	s.mu.Unlock()

	s.notify()
}

// Subscribe returns a channel that receives a tick on every Reload.
// Caller must call the returned cancel function when done.
func (s *Store) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	s.subsMu.Lock()
	s.subs[ch] = struct{}{}
	s.subsMu.Unlock()

	return ch, func() {
		s.subsMu.Lock()

		if _, ok := s.subs[ch]; ok {
			delete(s.subs, ch)
			close(ch)
		}

		s.subsMu.Unlock()
	}
}

func (s *Store) notify() {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()

	for ch := range s.subs {
		// Non-blocking send:
		// each subscriber has a buffer of 1,
		// if it's already full, the subscriber will pick up the latest version on its next read anyway.
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
