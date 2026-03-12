package watcher

import (
	"testing"
)

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

func TestFilterGitIgnoredFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	_ = filter.ShouldIgnore("/test/repo/somefile.go")
}

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
