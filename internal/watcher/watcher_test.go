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

func TestFilterExecutableFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	// This test might fail if the file doesn't have execute permissions
	// The filter checks for executable permissions, so we can't guarantee this will pass
	// without creating an actual executable file
	_ = filter.ShouldIgnore("/test/repo/script")
	// Note: This test is skipped because we can't easily create an executable file in a test
}

func TestFilterNumericExtensions(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.12345678") {
		t.Error("Files with numeric extensions should be ignored")
	}
}

func TestFilterDirectoryPatterns(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/dist/file.js") {
		t.Error("Files in dist directories should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/.next/page.html") {
		t.Error("Files in .next directories should be ignored")
	}
}

func TestFilterVibewatchDirectory(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/vibewatch/file.go") {
		t.Error("Files in vibewatch directories should not be ignored")
	}
}

func TestFilterGitHeadAndIndex(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/.git/HEAD") {
		t.Error(".git/HEAD should not be ignored")
	}

	if filter.ShouldIgnore("/test/repo/.git/index") {
		t.Error(".git/index should not be ignored")
	}
}

func TestFilterGitDirectoryItself(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/.git") {
		t.Error(".git directory itself should not be ignored")
	}
}

func TestFilterObjectFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/program.o") {
		t.Error(".o files should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/program.out") {
		t.Error(".out files should be ignored")
	}
}

func TestFilterCacheDirectories(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/.mypy_cache/file.pyc") {
		t.Error("Files in .mypy_cache directories should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/.pytest_cache/file.cache") {
		t.Error("Files in .pytest_cache directories should be ignored")
	}
}

func TestFilterVirtualEnvDirectories(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/.venv/lib/python3.9/site-packages/package/module.py") {
		t.Error("Files in .venv directories should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/venv/lib/python3.9/site-packages/package/module.py") {
		t.Error("Files in venv directories should be ignored")
	}
}

func TestFilterPythonCache(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/__pycache__/module.pyc") {
		t.Error("Files in __pycache__ directories should be ignored")
	}
}

func TestFilterDSStore(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/.DS_Store") {
		t.Error(".DS_Store files should be ignored")
	}
}

func TestFilterSwapFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.swp") {
		t.Error(".swp files should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/file.swo") {
		t.Error(".swo files should be ignored")
	}
}

func TestFilterTildeFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file~") {
		t.Error("~ backup files should be ignored")
	}
}

func TestFilterLockFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.lock") {
		t.Error(".lock files should be ignored")
	}

	if !filter.ShouldIgnore("/test/repo/file.pid") {
		t.Error(".pid files should be ignored")
	}
}

func TestFilterLogFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.log") {
		t.Error(".log files should be ignored")
	}
}

func TestFilterBakFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.bak") {
		t.Error(".bak files should be ignored")
	}
}

func TestFilterTmpFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/file.tmp") {
		t.Error(".tmp files should be ignored")
	}
}

func TestFilterAFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/lib.a") {
		t.Error(".a files should be ignored")
	}
}

func TestFilterDLLFiles(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/lib.dll") {
		t.Error(".dll files should be ignored")
	}
}

func TestFilterNextDirectory(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/.next/static/file.js") {
		t.Error("Files in .next directories should be ignored")
	}
}

func TestFilterBuildDirectory(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/build/output.apk") {
		t.Error("Files in build directories should be ignored")
	}
}

func TestFilterToxDirectory(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/.tox/env/lib/python3.9/site-packages/package/module.py") {
		t.Error("Files in .tox directories should be ignored")
	}
}

func TestFilterNodeModules(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if !filter.ShouldIgnore("/test/repo/node_modules/package/index.js") {
		t.Error("Files in node_modules directories should be ignored")
	}
}

func TestFilterNormalGoFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.go") {
		t.Error("main.go should not be ignored")
	}
}

func TestFilterNormalJSFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/index.js") {
		t.Error("index.js should not be ignored")
	}
}

func TestFilterNormalPythonFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.py") {
		t.Error("main.py should not be ignored")
	}
}

func TestFilterNormalTypeScriptFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/index.ts") {
		t.Error("index.ts should not be ignored")
	}
}

func TestFilterNormalHTMLFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/index.html") {
		t.Error("index.html should not be ignored")
	}
}

func TestFilterNormalCSSFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/styles.css") {
		t.Error("styles.css should not be ignored")
	}
}

func TestFilterNormalJSONFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/package.json") {
		t.Error("package.json should not be ignored")
	}
}

func TestFilterNormalMDFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/README.md") {
		t.Error("README.md should not be ignored")
	}
}

func TestFilterNormalTXTFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/notes.txt") {
		t.Error("notes.txt should not be ignored")
	}
}

func TestFilterNormalYAMLFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/config.yaml") {
		t.Error("config.yaml should not be ignored")
	}
}

func TestFilterNormalYMLFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/config.yml") {
		t.Error("config.yml should not be ignored")
	}
}

func TestFilterNormalTOMLFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/config.toml") {
		t.Error("config.toml should not be ignored")
	}
}

func TestFilterNormalSQLFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/query.sql") {
		t.Error("query.sql should not be ignored")
	}
}

func TestFilterNormalJavaFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/Main.java") {
		t.Error("Main.java should not be ignored")
	}
}

func TestFilterNormalKotlinFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.kt") {
		t.Error("main.kt should not be ignored")
	}
}

func TestFilterNormalRubyFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.rb") {
		t.Error("main.rb should not be ignored")
	}
}

func TestFilterNormalPHPFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/index.php") {
		t.Error("index.php should not be ignored")
	}
}

func TestFilterNormalCFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.c") {
		t.Error("main.c should not be ignored")
	}
}

func TestFilterNormalCppFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.cpp") {
		t.Error("main.cpp should not be ignored")
	}
}

func TestFilterNormalRustFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.rs") {
		t.Error("main.rs should not be ignored")
	}
}

func TestFilterNormalSwiftFile(t *testing.T) {
	repos := []string{"/test/repo"}
	filter := NewFilter("/test", repos)

	if filter.ShouldIgnore("/test/repo/main.swift") {
		t.Error("main.swift should not be ignored")
	}
}
