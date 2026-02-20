package types

import "time"

// DiffEntry represents a single observed file change with its computed diff.
type DiffEntry struct {
	FilePath  string
	Repo      string // repo name (directory basename), empty for single-repo mode
	Timestamp time.Time
	Diff      string // raw unified diff text
	IsNew     bool
	IsDeleted bool
	Error     string // non-fatal error message
}
