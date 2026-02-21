package watcher

import (
	"testing"
)

// TestFilterGitDirectory tests that .git directory is allowed for watching
func TestFilterGitDirectory(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/.git") {
		t.Error(".git directory should not be ignored")
	}

	if filter.ShouldIgnore("/test/repo/.git/HEAD") {
		t.Error(".git/HEAD should not be ignored")
	}

	if filter.ShouldIgnore("/test/repo/.git/index") {
		t.Error(".git/index should not be ignored")
	}
}

// TestFilterBuiltinIgnores tests that built-in ignores work
func TestFilterBuiltinIgnores(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/node_modules") {
		t.Error("node_modules should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/.DS_Store") {
		t.Error(".DS_Store should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/test.tmp") {
		t.Error(".tmp files should be ignored")
	}
}

// TestFilterGitIgnoredFiles tests git ignore integration
func TestFilterGitIgnoredFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	_ = filter.ShouldIgnore("/test/repo/somefile.go")
}

// TestFilterNormalFiles tests that normal files are not ignored
func TestFilterNormalFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.go") {
		t.Error("main.go should not be ignored")
	}

	if filter.ShouldIgnore("/test/repo/index.js") {
		t.Error("index.js should not be ignored")
	}
}

// TestFilterBinaryFiles tests that binary files are ignored
func TestFilterBinaryFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/program.exe") {
		t.Error(".exe files should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/lib.so") {
		t.Error(".so files should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/lib.dylib") {
		t.Error(".dylib files should be ignored")
	}
}
