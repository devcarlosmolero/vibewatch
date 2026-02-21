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

// initWatcherDebugLogging initializes the debug log file for the watcher
func initWatcherDebugLogging(root string) {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile != nil {
		debugFile.Close()
	}

	debugPath := fmt.Sprintf("%s/watch.log", root)
	var err error
	debugFile, err = os.Create(debugPath)
	if err != nil {
		debugFile = nil
		return
	}

	debugFile.WriteString("=== Watcher Debug Log ===\n")
	debugFile.WriteString(fmt.Sprintf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	debugFile.Sync()
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

	initWatcherDebugLogging(root)

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if filter.ShouldIgnore(path) {
				return filepath.SkipDir
			}
			if addErr := fsw.Add(path); addErr != nil {
				return nil
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

			paths := make([]string, 0, len(w.pending))
			for path := range w.pending {
				paths = append(paths, path)
			}
			w.pending = make(map[string]struct{})
			w.pendingMu.Unlock()

			logMessage(fmt.Sprintf("Processing batch of %d changes", len(paths)))
			for _, path := range paths {
				select {
				case w.changes <- path:
					if path == "__GIT_OPERATION__" {
						logMessage("Sent git operation marker to channel")
					}
				case <-w.done:
					return
				}
			}
		})
	} else {
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
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	if w.filter.ShouldIgnore(path) {
		return
	}

	if event.Has(fsnotify.Create) {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			w.fsw.Add(path)
			return
		}
	}

	if strings.Contains(path, ".git") && (filepath.Base(path) == "HEAD" || filepath.Base(path) == "index") {
		logMessage(fmt.Sprintf("Detected git operation (%s changed), triggering full refresh", filepath.Base(path)))
		w.pendingMu.Lock()
		w.pending["__GIT_OPERATION__"] = struct{}{}
		w.pendingMu.Unlock()
		w.scheduleBatch()
		return
	}

	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) &&
		!event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
		return
	}

	w.pendingMu.Lock()
	w.pending[path] = struct{}{}
	w.pendingMu.Unlock()

	w.scheduleBatch()
}
