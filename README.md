# Vibewatch

[![Tests](https://woodpecker.carlosmolero.com/api/badges/11/status.svg)](https://woodpecker.carlosmolero.com/repos/11)

**Vibewatch** is a terminal-based tool that helps developers track code changes in real-time when working with CLI AI agents. It provides instant visibility into file modifications, additions, and deletions, making it ideal for monitoring and understanding changes made by AI coding assistants to your codebase.

<a href="https://codeberg.org/devcarlosmolero/vibewatch" target="_blank" noopener noreferrer><img width="200px" src="https://codeberg.org/devcarlosmolero/vibewatch/raw/branch/master/mirror.svg"></a>

## Table of Contents

- [Key Features](#key-features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [Advanced Options](#advanced-options)
- [How It Works](#how-it-works)
- [Debugging](#debugging)
- [License](#license)

## Key Features

- **Real-time file monitoring**: See changes as they happen in your repository
- **Git diff integration**: View actual code changes, not just file names
- **Multi-repository support**: Monitor multiple Git repositories simultaneously
- **Interactive TUI**: Clean, keyboard-navigable terminal interface
- **AI agent friendly**: Designed to help track changes made by CLI AI coding assistants
- **Batch processing**: Efficiently handles rapid file changes

## Requirements

- Go 1.21+
- Git

## Installation

```bash
brew tap devcarlosmolero/homebrew https://codeberg.org/devcarlosmolero/homebrew
brew install devcarlosmolero/homebrew/vibewatch
```

Or build from source:

```bash
git clone https://codeberg.org/devcarlosmolero/vibewatch.git
cd vibewatch
go build -o vibewatch .
```

## Usage

### Basic Usage

Start monitoring the current directory (must be a Git repository):

```bash
vibewatch
```

Monitor a specific repository:

```bash
vibewatch -dir /path/to/your/repo
```

### Advanced Options

| Flag         | Description                                                                                                                                           |
| :----------- | :---------------------------------------------------------------------------------------------------------------------------------------------------- |
| **-dir**     | Specify the directory to watch. Defaults to current directory. Can be a single Git repository or a parent directory containing multiple repositories. |
| **-max**     | Set the maximum number of diff entries to keep (default: 200). Useful for limiting memory usage in large repositories.                                |
| **-version** | Print the version of Vibewatch and exit.                                                                                                              |

### Monitoring Multiple Repositories

Vibewatch can monitor directories containing multiple Git repositories:

```bash
vibewatch -dir /path/to/parent/directory
```

This will automatically detect and monitor all Git repositories within the specified directory.

### Keyboard Controls

- **Arrow keys**: Navigate through changes
- **q or Ctrl+C**: Quit the application
- **?**: Show help/keybindings

## How It Works

Vibewatch uses a sophisticated pipeline to monitor and display file changes:

1. **Filesystem Watching**: Uses Go's fsnotify to detect file changes
2. **Event Filtering**: Ignores irrelevant files (like .git directory, temporary files)
3. **Batch Processing**: Groups rapid changes together for efficiency
4. **Git Diff Computation**: Shows actual code changes for modified files
5. **TUI Rendering**: Displays changes in a clean, interactive terminal interface

The batch processing system is particularly important - it groups changes that occur within 100ms of each other, preventing UI overload during rapid file modifications.

## Debugging

Vibewatch creates detailed debug logs to help troubleshoot issues:

- **watcher.log**: Filesystem events and processing
- **model.log**: Diff computation and UI updates

## License

[![MIT License][mit-shield]][mit]

This work is licensed under the [MIT License][mit].

[mit]: https://opensource.org/licenses/MIT
[mit-image]: https://img.shields.io/badge/License-MIT-yellow.svg
[mit-shield]: https://img.shields.io/badge/License-MIT-lightgrey.svg
