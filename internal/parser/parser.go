// Package parser extracts structured goroutine data from raw pprof
// goroutine dump text (debug=2 format).
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Goroutine represents a single parsed goroutine from a dump.
type Goroutine struct {
	ID     int     // goroutine ID
	State  string  // e.g. "running", "IO wait", "chan receive"
	Frames []Frame // stack frames, top of stack first
}

// Frame represents a single stack frame within a goroutine.
type Frame struct {
	Function string // fully qualified function name, e.g. "net/http.(*connReader).Read"
	File     string // source file path, e.g. "/usr/local/go/src/net/http/server.go"
	Line     int    // line number in the source file
}

// headerRe matches the goroutine header line:
//
//	goroutine 6 [IO wait]:
//	goroutine 1 [running]:
//	goroutine 9 [chan receive, 5 minutes]:
var headerRe = regexp.MustCompile(`^goroutine\s+(\d+)\s+\[([^\]]+)\]:`)

// fileLineRe matches the file:line reference line (with leading tab):
//
//	\t/usr/local/go/src/net/http/server.go:786 +0x158
var fileLineRe = regexp.MustCompile(`^\t(.+):(\d+)\s`)

// createdByRe matches the "created by" line format:
//
//	created by main.startWorkers in goroutine 1
var createdByRe = regexp.MustCompile(`^created by\s+(\S+)`)

// extractFuncName extracts the fully qualified function name from a
// function call line, handling method receivers like (*Type).Method.
//
// Examples:
//
//	"net/http.(*connReader).Read(0xc...)" → "net/http.(*connReader).Read"
//	"main.worker(0xc000100060)"           → "main.worker"
//	"created by main.startWorkers in goroutine 1" → "main.startWorkers"
func extractFuncName(line string) string {
	// Handle "created by" lines
	if matches := createdByRe.FindStringSubmatch(line); matches != nil {
		return matches[1]
	}

	// For regular function lines, find the argument list by scanning
	// for the outermost '(' that starts the arguments.
	// We need to skip '(' inside receiver types like (*connReader).
	depth := 0
	for i, ch := range line {
		switch ch {
		case '(':
			if depth == 0 && i > 0 {
				// Check if this looks like a receiver: preceded by a dot
				// or is at the start — if the char before is '.', this is
				// a receiver like (*Type), not the argument list.
				// The argument list '(' is the one NOT preceded by a '.'
				// after a ')' or at a package boundary.
				name := strings.TrimSpace(line[:i])
				if len(name) > 0 && name[len(name)-1] == '.' {
					// This is a receiver start: pkg.(*Type)
					depth++
					continue
				}
				// This is the argument list start
				return name
			}
			depth++
		case ')':
			depth--
		}
	}

	// Fallback: return the whole line trimmed
	return strings.TrimSpace(line)
}

// Parse parses a raw goroutine dump string (pprof debug=2 format) and
// returns a slice of structured Goroutine objects.
//
// The expected format for each goroutine block is:
//
//	goroutine <id> [<state>]:
//	<function name>(<args>)
//		<file>:<line> +<offset>
//	...
//
// Blocks are separated by blank lines.
func Parse(input string) ([]Goroutine, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("empty input: no goroutine data to parse")
	}

	var goroutines []Goroutine
	var current *Goroutine
	var pendingFunc string

	lines := strings.Split(input, "\n")

	for _, line := range lines {
		// Check for goroutine header
		if matches := headerRe.FindStringSubmatch(line); matches != nil {
			// Save previous goroutine if any
			if current != nil {
				goroutines = append(goroutines, *current)
			}

			id, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, fmt.Errorf("invalid goroutine id %q: %w", matches[1], err)
			}

			// Strip duration info from state (e.g. "chan receive, 5 minutes" → "chan receive")
			state := matches[2]
			if idx := strings.Index(state, ","); idx != -1 {
				state = strings.TrimSpace(state[:idx])
			}

			current = &Goroutine{
				ID:    id,
				State: state,
			}
			pendingFunc = ""
			continue
		}

		// Skip lines before the first goroutine header
		if current == nil {
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Blank line — end of current goroutine block
		if trimmed == "" {
			if current != nil {
				goroutines = append(goroutines, *current)
				current = nil
				pendingFunc = ""
			}
			continue
		}

		// File:line reference line (starts with tab)
		if matches := fileLineRe.FindStringSubmatch(line); matches != nil && strings.HasPrefix(line, "\t") {
			lineNum, err := strconv.Atoi(matches[2])
			if err != nil {
				return nil, fmt.Errorf("invalid line number %q in goroutine %d: %w", matches[2], current.ID, err)
			}

			funcName := pendingFunc
			if funcName == "" {
				funcName = "<unknown>"
			}

			current.Frames = append(current.Frames, Frame{
				Function: funcName,
				File:     matches[1],
				Line:     lineNum,
			})
			pendingFunc = ""
			continue
		}

		// Function name line (does not start with tab)
		if !strings.HasPrefix(line, "\t") {
			pendingFunc = extractFuncName(line)
			continue
		}
	}

	// Don't forget the last goroutine if input doesn't end with blank line
	if current != nil {
		goroutines = append(goroutines, *current)
	}

	if len(goroutines) == 0 {
		return nil, fmt.Errorf("no goroutines found in input")
	}

	return goroutines, nil
}
