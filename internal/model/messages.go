package model

import "codeberg.org/devcarlosmolero/vibewatch/internal/types"

// FileChangedMsg is sent when a file change is detected and diffed.
type FileChangedMsg types.DiffEntry

// InitialEntriesMsg carries pre-existing dirty files found at startup.
type InitialEntriesMsg []types.DiffEntry

// ToggleFileMsg is sent when a file's visibility should be toggled.
type ToggleFileMsg string

// ShowAllFilesMsg is sent when all files should be made visible.
type ShowAllFilesMsg bool

// UntoggleFileMsg is sent when a specific file should be made visible.
type UntoggleFileMsg string
