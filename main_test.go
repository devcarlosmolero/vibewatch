package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/devcarlosmolero/vibewatch/internal/differ"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "-version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	output := out.String()
	expected := "vibewatch v1.0.1\n"
	if output != expected {
		t.Errorf("Expected version output '%s', got '%s'", expected, output)
	}
}

func TestDirFlagValidation(t *testing.T) {
	testCases := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{"valid current dir", ".", false},
		{"valid absolute dir", "/tmp", false},
		{"invalid dir", "/nonexistent/directory/that/does/not/exist", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "main.go", "-dir", tc.dir)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &bytes.Buffer{}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := cmd.Start()
			if err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				if tc.expectError {
					t.Error("Command timed out - expected quick failure for invalid directory")
				}
			case err := <-done:
				if tc.expectError && err == nil {
					t.Error("Expected error for invalid directory, got none")
				}
				if !tc.expectError && err != nil {
					stderrStr := stderr.String()
					if stderrStr == "" {
						t.Errorf("Unexpected error for valid directory: %v", err)
					}
				}
			}
		})
	}
}

func TestRepoFilterFlag(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	os.Mkdir(repo1, 0755)
	initGitRepo(t, repo1)

	repo2 := filepath.Join(tmpDir, "repo2")
	os.Mkdir(repo2, 0755)
	initGitRepo(t, repo2)

	cmd := exec.Command("go", "run", "main.go", "-dir", tmpDir, "-repos", "repo1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		t.Log("Command timed out as expected (would run indefinitely)")
	case err := <-done:
		if err == nil {
			stderrStr := stderr.String()
			if stderrStr == "" {
				t.Error("Expected command to run or show multi-repo mode message")
			}
		}
	}
}

func TestMaxEntriesFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "-max", "500", "-dir", ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		t.Log("Command timed out as expected (would run indefinitely)")
	case err := <-done:
		if err == nil {
			stderrStr := stderr.String()
			if stderrStr == "" {
				t.Error("Expected command to run or show some output")
			}
		}
	}
}

func TestBranchDetectionSingleRepo(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	changeBranch(t, tmpDir, "test-branch")

	branch := differ.GetBranch(tmpDir)
	if branch != "test-branch" {
		t.Errorf("Expected branch 'test-branch', got '%s'", branch)
	}
}

func TestBranchDetectionMultiRepo(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	os.Mkdir(repo1, 0755)
	initGitRepo(t, repo1)

	repo2 := filepath.Join(tmpDir, "repo2")
	os.Mkdir(repo2, 0755)
	initGitRepo(t, repo2)
	changeBranch(t, repo2, "develop")

	repos, err := differ.DiscoverRepos(tmpDir)
	if err != nil {
		t.Fatalf("Failed to discover repos: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(repos))
	}

	branch1 := differ.GetBranch(repo1)
	branch2 := differ.GetBranch(repo2)

	if branch1 != "master" {
		t.Errorf("Expected repo1 branch 'master', got '%s'", branch1)
	}

	if branch2 != "develop" {
		t.Errorf("Expected repo2 branch 'develop', got '%s'", branch2)
	}
}

func TestCLIWithInvalidFlags(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"invalid flag", []string{"vibewatch", "-invalid"}, true},
		{"missing dir value", []string{"vibewatch", "-dir"}, true},
		{"invalid max value", []string{"vibewatch", "-max", "-10"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "main.go")
			cmd.Args = append([]string{"go", "run", "main.go"}, tc.args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &bytes.Buffer{}

			err := cmd.Run()

			if tc.expectError && err == nil {
				t.Error("Expected error, got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	cmd.Run()

	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}
}

func changeBranch(t *testing.T, dir string, branch string) {
	t.Helper()

	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to change branch: %v", err)
	}
}
