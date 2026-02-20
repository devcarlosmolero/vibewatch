package main


import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"codeberg.org/devcarlosmolero/vibewatch/internal/differ"
	"codeberg.org/devcarlosmolero/vibewatch/internal/model"
	"codeberg.org/devcarlosmolero/vibewatch/internal/watcher"
)

func main() {
	dir := flag.String("dir", ".", "directory to watch (git repo or parent of multiple repos)")
	maxEntries := flag.Int("max", 200, "maximum number of diff entries to keep")
	flag.Parse()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a valid directory\n", absDir)
		os.Exit(1)
	}

	var d differ.Differ
	var modeLabel string
	var repoRoots []string
	var repoNames []string
	var branches map[string]string
	var singleBranch string

	if differ.IsGitRepo(absDir) {
		// Single repo mode
		gd, err := differ.NewGit(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		d = gd
		modeLabel = absDir
		repoRoots = []string{absDir}
		repoNames = []string{filepath.Base(absDir)}
		singleBranch = differ.GetBranch(absDir)
	} else {
		// Multi-repo mode: discover git repos inside this directory
		repos, err := differ.DiscoverRepos(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning for repos: %v\n", err)
			os.Exit(1)
		}
		if len(repos) == 0 {
			fmt.Fprintf(os.Stderr, "Error: %s is not a git repository and contains no git repositories.\n", absDir)
			fmt.Fprintf(os.Stderr, "Point vibewatch at a git repo or a directory containing repos.\n")
			os.Exit(1)
		}

		md, err := differ.NewMulti(repos)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		d = md
		repoRoots = md.RepoRoots()

		repoNames = make([]string, 0, len(repos))
		branches = make(map[string]string)
		for root, name := range repos {
			repoNames = append(repoNames, name)
			branches[name] = differ.GetBranch(root)
		}
		sort.Strings(repoNames)
		modeLabel = fmt.Sprintf("%s (%d repos)", absDir, len(repos))
		fmt.Fprintf(os.Stderr, "Multi-repo mode: watching %d repositories\n", len(repos))
		for root, name := range repos {
			fmt.Fprintf(os.Stderr, "  %s (%s)\n", name, root)
		}
	}

	// Create filter and watcher
	filter := watcher.NewFilter(absDir, repoRoots)
	w, err := watcher.New(absDir, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting watcher: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()

	// Run TUI with context for graceful shutdown
	m := model.New(w.Changes(), d, *maxEntries, modeLabel, repoNames, branches, singleBranch)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseAllMotion(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// Test comment for watching - this should trigger a change detection
// Another test comment to verify real-time updates are working
// Third test comment - checking if multiple rapid changes work correctly
// Fourth test - after fixing the processPendingChanges function
// Fifth test - checking if the watcher is properly detecting changes now
