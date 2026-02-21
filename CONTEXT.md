# Vibewatch Architecture Documentation

This document explains how Vibewatch's real-time file watching system works, including the recent fixes that made it functional.

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [File Watching Pipeline](#file-watching-pipeline)
4. [The Batch Processing Fix](#the-batch-processing-fix)
5. [Debugging Guide](#debugging-guide)
6. [Performance Considerations](#performance-considerations)

## System Overview

Vibewatch is a real-time file change monitoring tool that detects modifications in a git repository and displays them in a TUI (Terminal User Interface). It uses Go's fsnotify library for filesystem notifications and Bubble Tea for the UI.

## Component Architecture

### 1. Watcher (`internal/watcher/`)

The watcher is responsible for:
- Setting up filesystem watches on directories
- Receiving fsnotify events
- Filtering out unwanted files
- Batching rapid changes together
- Sending changes to the model via a channel

**Key Files:**
- `watcher.go` - Main watcher logic and event handling
- `filter.go` - File filtering logic

### 2. Model (`internal/model/`)

The model handles:
- Receiving changes from the watcher
- Computing git diffs for changed files
- Managing the TUI state
- Rendering the viewport

**Key Files:**
- `model.go` - Main model logic and Bubble Tea integration
- `styles.go` - UI styling
- `help.go` - Help system

### 3. Differ (`internal/differ/`)

The differ computes git diffs for changed files:
- Finds the git repository root
- Executes git commands to get diffs
- Handles various git diff scenarios

### 4. Main Application (`main.go`)

- Initializes all components
- Sets up the Bubble Tea program
- Handles command-line arguments

## File Watching Pipeline

Here's how a file change flows through the system:

```
File System Change
        ↓
fsnotify detects event
        ↓
Watcher.handleEvent() receives event
        ↓
Filter.ShouldIgnore() checks if file should be ignored
        ↓
Event added to w.pending map
        ↓
scheduleBatch() called
        ↓
  If no timer: Create new timer (100ms)
  If timer exists: Reset timer (100ms) ← CRITICAL FIX
        ↓
Timer fires after 100ms
        ↓
Batch processing callback executes
        ↓
Pending changes sent to w.changes channel
        ↓
Model receives FileChangedMsg
        ↓
Model computes git diff
        ↓
Model updates entries
        ↓
Viewport content updated
        ↓
TUI redraws with new changes
```

## The Batch Processing Fix

### The Problem

The original `scheduleBatch()` function had a critical bug:

```go
func (w *Watcher) scheduleBatch() {
    if w.batchTimer == nil {
        // Create timer...
    }
    // MISSING: else clause to reset existing timer!
}
```

**What happened:**
1. First change → Timer created (would fire after 100ms) ✅
2. Second change within 100ms → No timer reset, so batch never scheduled ❌
3. Timer fires → Only processes first change
4. Subsequent changes → Never processed ❌

### The Solution

Added the missing timer reset logic:

```go
func (w *Watcher) scheduleBatch() {
    if w.batchTimer == nil {
        // Create new timer
        w.batchTimer = time.AfterFunc(batchInterval, func() {
            // Process batch...
        })
    } else {
        // Reset existing timer ← THIS FIXES IT!
        w.batchTimer.Reset(batchInterval)
    }
}
```

**Why it works now:**
- Each change resets the 100ms timer
- All changes within 100ms are batched together
- Timer callback executes and sends all changes to the model
- Model receives and displays changes in real-time

## Debugging Guide

### Debug Logs

Vibewatch creates two debug log files:

1. **`vibewatch_debug.log`** - Watcher events and processing
2. **`model.log`** - Model processing and diff computation

**What to look for:**

**In watcher log:**
- `Raw event received: /path/to/file` - fsnotify detected a change
- `Detected WRITE event for: /path/to/file` - Event passed filtering
- `Processing batch of N changes` - Batch processing started
- `Successfully sent to channel: /path/to/file` - Change sent to model

**In model log:**
- `MODEL: Received change from channel: /path/to/file` - Model received change
- `MODEL: Successfully got diff for /path/to/file` - Diff computed successfully

### Common Issues and Solutions

**Issue: Changes not appearing in TUI**
- Check if watcher log shows `Processing batch` messages
- If not, batch timer isn't firing (check scheduleBatch logic)
- Check if model log shows received messages
- If not, channel communication is broken

**Issue: Too many unwanted files shown**
- Adjust filtering logic in `filter.go`
- Check `.gitignore` patterns
- Add more file extensions to ignore list

**Issue: Performance lag with many changes**
- Increase `batchInterval` (currently 100ms)
- Reduce `maxBatchSize` (currently 50)
- Optimize git diff computation

## Performance Considerations

### Batch Processing

- **Batch Interval:** 100ms (adjustable)
- **Max Batch Size:** 50 changes per batch
- **Trade-off:** Smaller batches = more responsive but more overhead

### Filesystem Watching

- fsnotify watches entire directory trees
- Each directory requires a filesystem watch handle
- Some filesystems have limits on watch handles

### Git Diff Computation

- Most expensive operation in the pipeline
- Cached to avoid recomputation
- Consider adding debounce for very active files

## Development Tips

### Adding New Features

1. **New file types:** Update filter logic in `filter.go`
2. **New UI elements:** Modify model and viewport rendering
3. **New commands:** Add to model's `Update()` method

### Testing Changes

1. Make a file change
2. Check debug logs for processing
3. Verify TUI updates
4. Test edge cases (rapid changes, large files, etc.)

### Code Organization

- Keep watcher logic separate from UI logic
- Maintain clean separation between components
- Add comprehensive logging for debugging

## Future Enhancements

### Potential Features

- **Multi-repository support:** Watch multiple git repos simultaneously
- **Advanced filtering:** Regex patterns, custom ignore rules
- **Historical view:** Scroll through past changes
- **Statistics:** Change frequency, most active files
- **Notifications:** Desktop notifications for important changes

### Performance Optimizations

- **Smart batching:** Adaptive batch intervals based on change frequency
- **Parallel diffs:** Compute multiple diffs concurrently
- **Lazy loading:** Only compute diffs for visible files
- **Memory optimization:** Better caching strategies

## Summary

Vibewatch provides real-time monitoring of git repository changes with a clean TUI interface. The system uses:

- **fsnotify** for filesystem event detection
- **Bubble Tea** for terminal UI
- **Git commands** for diff computation
- **Batch processing** for handling rapid changes

The recent fix to the batch timer reset logic was critical for making real-time updates work correctly. With this fix in place, the system now properly batches and processes file changes, providing the real-time monitoring experience that was originally intended.

## Troubleshooting Checklist

1. **No changes appearing?**
   - ✅ Check watcher debug log for event detection
   - ✅ Check for "Processing batch" messages
   - ✅ Check model debug log for message reception
   - ✅ Verify channel communication is working

2. **Too many unwanted files?**
   - ✅ Review filter logic in `filter.go`
   - ✅ Check `.gitignore` patterns
   - ✅ Adjust ignored extensions list

3. **Performance issues?**
   - ✅ Try increasing batch interval
   - ✅ Reduce max batch size
   - ✅ Check for expensive git operations

4. **Crashes or errors?**
   - ✅ Check debug logs for error messages
   - ✅ Test with simpler repository
   - ✅ Verify filesystem permissions

This architecture provides a solid foundation for real-time file monitoring that can be extended with additional features as needed.