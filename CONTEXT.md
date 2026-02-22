# Vibewatch Architecture Documentation

This document explains how Vibewatch's real-time file watching system works.

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
If timer exists: Reset timer (100ms)
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

This architecture provides a solid foundation for real-time file monitoring that can be extended with additional features as needed.

