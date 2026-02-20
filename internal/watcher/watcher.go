package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	dbounceInterval = 300 * time.Millisecond
	batchInterval   = 100 * time.Millisecond
	maxBatchSize    = 50
)

var (
	debugFile  *os.File
	debugMutex sync.Mutex
)

// logMessage writes a debug message to the debug file
func logMessage(message string) {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile != nil {
		timestamp := time.Now().Format("15:04:05")
		debugFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
		debugFile.Sync()
	}
}

// Watcher monitors a directory recursively for file changes.
type Watcher struct {
	root       string
	filter     *Filter
	fsw        *fsnotify.Watcher
	changes    chan string
	pending    map[string]struct{}
	batchTimer *time.Timer
	pendingMu  sync.Mutex
	done       chan struct{}
}

// New creates a recursive file watcher on the given root directory.
func New(root string, filter *Filter) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		root:    root,
		filter:  filter,
		fsw:     fsw,
		changes: make(chan string, 64),
		pending: make(map[string]struct{}),
		done:    make(chan struct{}),
	}

	// Walk and add all directories
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if d.IsDir() {
			if filter.ShouldIgnore(path) {
				return filepath.SkipDir
			}
			if addErr := fsw.Add(path); addErr != nil {
				return nil // skip directories that can't be watched (broken symlinks, etc.)
			}
		}
		return nil
	})
	if err != nil {
		fsw.Close()
		return nil, err
	}

	go w.loop()
	return w, nil
}

// Changes returns a read-only channel that emits changed file paths.
func (w *Watcher) Changes() <-chan string {
	return w.changes
}

// Close stops the watcher and releases resources.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}

func (w *Watcher) scheduleBatch() {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	if w.batchTimer == nil {
		w.batchTimer = time.AfterFunc(batchInterval, func() {
			w.pendingMu.Lock()
			if len(w.pending) == 0 {
				w.pendingMu.Unlock()
				return
			}

			// Create a slice of pending paths
			paths := make([]string, 0, len(w.pending))
			for path := range w.pending {
				paths = append(paths, path)
			}
			w.pending = make(map[string]struct{})
			w.pendingMu.Unlock()

			// Send all paths in the batch
			logMessage(fmt.Sprintf("Processing batch of %d changes", len(paths)))
			for _, path := range paths {
				select {
				case w.changes <- path:
					// Path sent successfully
					if path == "__GIT_OPERATION__" {
						logMessage("Sent git operation marker to channel")
					}
				case <-w.done:
					return
				}
			}
		})
	} else {
		// Reset the timer to extend the batch interval
		w.batchTimer.Reset(batchInterval)
	}
}

func (w *Watcher) loop() {
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			// silently ignore watcher errors
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	if w.filter.ShouldIgnore(path) {
		return
	}

	// If a new directory was created, start watching it
	if event.Has(fsnotify.Create) {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			w.fsw.Add(path)
			return
		}
	}

	// Special handling for git operations (commit, etc.)
	// When .git/HEAD or .git/index changes, we need to refresh all files
	if strings.Contains(path, ".git") && (filepath.Base(path) == "HEAD" || filepath.Base(path) == "index") {
		logMessage(fmt.Sprintf("Detected git operation (%s changed), triggering full refresh", filepath.Base(path)))
		// Send a special marker to indicate a git operation
		w.pendingMu.Lock()
		w.pending["__GIT_OPERATION__"] = struct{}{}
		w.pendingMu.Unlock()
		w.scheduleBatch()
		return
	}

	// Only care about write, create, remove, and rename events for files
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) &&
		!event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
		return
	}

	w.pendingMu.Lock()
	w.pending[path] = struct{}{}
	w.pendingMu.Unlock()

	// Schedule a batch if not already scheduled
	w.scheduleBatch()
}
