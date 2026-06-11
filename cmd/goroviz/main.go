// Goroviz — "htop for Go goroutines"
//
// Install:
//
//	go install github.com/santhosh-005/goroviz/cmd/goroviz@latest
//
// Usage:
//
//	goroviz attach <url>       Fetch goroutine dump from a live Go app
//	goroviz dump <file>        Read goroutine dump from a file
//
// Examples:
//
//	goroviz attach localhost:6060
//	goroviz attach http://myapp.example.com:8080
//	goroviz dump goroutines.txt
//	goroviz dump goroutines.txt --text
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-005/goroviz/internal/fetcher"
	"github.com/santhosh-005/goroviz/internal/group"
	"github.com/santhosh-005/goroviz/internal/parser"
	"github.com/santhosh-005/goroviz/internal/tui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	// Parse global flags
	textMode := false
	var filteredArgs []string
	for _, arg := range args {
		switch arg {
		case "--text", "-t":
			textMode = true
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) == 0 {
		printUsage()
		os.Exit(1)
	}

	subcommand := filteredArgs[0]

	switch subcommand {
	case "attach":
		if len(filteredArgs) < 2 {
			fmt.Fprintf(os.Stderr, "Error: missing URL\n\n")
			fmt.Fprintf(os.Stderr, "Usage: goroviz attach <url>\n")
			fmt.Fprintf(os.Stderr, "Example: goroviz attach localhost:6060\n")
			os.Exit(1)
		}
		url := filteredArgs[1]
		runAttach(url, textMode)

	case "dump":
		if len(filteredArgs) < 2 {
			fmt.Fprintf(os.Stderr, "Error: missing file path\n\n")
			fmt.Fprintf(os.Stderr, "Usage: goroviz dump <file>\n")
			fmt.Fprintf(os.Stderr, "Example: goroviz dump goroutines.txt\n")
			os.Exit(1)
		}
		filePath := filteredArgs[1]
		runDump(filePath, textMode)

	case "help":
		printUsage()
		os.Exit(0)

	default:
		// If it looks like a URL, treat it as "attach"
		if looksLikeURL(subcommand) {
			runAttach(subcommand, textMode)
		} else if fileExists(subcommand) {
			// If it's an existing file, treat as "dump"
			runDump(subcommand, textMode)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
			printUsage()
			os.Exit(1)
		}
	}
}

// runAttach fetches goroutine dump from a live app and launches the TUI.
func runAttach(url string, textMode bool) {
	fmt.Fprintf(os.Stderr, "⏳ Fetching goroutine dump from %s ...\n", url)

	data, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ Fetched goroutine dump\n\n")

	analyze(data, textMode)
}

// runDump reads a goroutine dump from file and launches the TUI.
func runDump(filePath string, textMode bool) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	analyze(string(data), textMode)
}

// analyze parses, groups, and displays goroutine data.
func analyze(data string, textMode bool) {
	goroutines, err := parser.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing goroutine dump: %v\n", err)
		os.Exit(1)
	}

	strategy := group.NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	if textMode {
		printSummary(goroutines, groups)
	} else {
		if err := tui.Run(groups); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
	}
}

// looksLikeURL checks if a string looks like a URL or host:port.
func looksLikeURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	// host:port pattern
	if strings.Contains(s, ":") && !strings.HasPrefix(s, "-") {
		return true
	}
	return false
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `🔍 Goroviz — htop for Go goroutines

Usage:
  goroviz attach <url>     Fetch from a running Go app's pprof endpoint
  goroviz dump <file>      Read goroutine dump from a file

Examples:
  goroviz attach localhost:6060
  goroviz attach http://myapp.example.com:8080
  goroviz dump goroutines.txt

Flags:
  -t, --text    Print plain text output (no TUI)
  -h, --help    Show this help message

Your Go app needs pprof enabled:
  import _ "net/http/pprof"
  go http.ListenAndServe("localhost:6060", nil)
`)
}

func printSummary(goroutines []parser.Goroutine, groups []group.Group) {
	fmt.Printf("Goroviz — Goroutine Analysis\n")
	fmt.Printf("%s\n\n", strings.Repeat("─", 60))
	fmt.Printf("Total goroutines: %d\n", len(goroutines))
	fmt.Printf("Unique groups:    %d\n\n", len(groups))
	fmt.Printf("%-6s  %-12s  %-s\n", "COUNT", "STATE", "TOP FUNCTION")
	fmt.Printf("%s\n", strings.Repeat("─", 60))

	for _, g := range groups {
		fmt.Printf("%-6d  %-12s  %s\n", g.Count(), g.State, g.TopFunction())
	}

	fmt.Printf("%s\n", strings.Repeat("─", 60))

	fmt.Printf("\n")
	for i, g := range groups {
		fmt.Printf("── Group %d: %s (%d goroutines) [%s] ──\n\n",
			i+1, g.TopFunction(), g.Count(), g.State)

		ids := make([]string, len(g.Goroutines))
		for j, gr := range g.Goroutines {
			ids[j] = fmt.Sprintf("%d", gr.ID)
		}
		fmt.Printf("  Goroutine IDs: %s\n\n", strings.Join(ids, ", "))

		fmt.Printf("  Stack trace:\n")
		for _, frame := range g.Goroutines[0].Frames {
			fmt.Printf("    %s\n", frame.Function)
			fmt.Printf("        %s:%d\n", frame.File, frame.Line)
		}
		fmt.Printf("\n")
	}
}
