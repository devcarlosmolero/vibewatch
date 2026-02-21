# Vibewatch Agent Documentation

This file provides information for agents working with the Vibewatch codebase.

## Architecture Documentation

For complete architecture documentation, including:
- System overview and component architecture
- File watching pipeline explanation
- Batch processing details
- Debugging guide and performance considerations
- Development tips and future enhancements

Please refer to the **[CONTEXT.md](CONTEXT.md)** file.

## Agent-Specific Information

### Code Standards

- **Commenting:** Only function/class-level documentation comments are used
- **Inline comments:** Minimal, only when necessary to explain WHY (not WHAT)
- **Code style:** Follow existing patterns in the codebase
- **Testing:** All changes should maintain existing test coverage

### Development Workflow

1. **Understand the architecture** by reading CONTEXT.md
2. **Run tests** before making changes: `go test ./...`
3. **Maintain clean code** with proper documentation
4. **Update tests** if functionality changes
5. **Verify builds** work: `go build -o vibewatch .`

### Key Components

- **Watcher** (`internal/watcher/`): Filesystem monitoring
- **Model** (`internal/model/`): TUI state management
- **Differ** (`internal/differ/`): Git diff computation
- **Main** (`main.go`): Application entry point

### Debugging

Debug logs are created in the repository root:
- `vibewatch_debug.log` - Watcher events
- `model_debug.log` - Model processing

See CONTEXT.md for detailed debugging information.
