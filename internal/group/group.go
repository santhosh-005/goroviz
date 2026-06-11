// Package group provides goroutine grouping functionality.
// It defines data types for grouped goroutines and a pluggable
// Strategy interface for different grouping algorithms.
package group

import (
	"github.com/santhosh-005/goroviz/internal/parser"
)

// Group represents a collection of goroutines that share a common
// stack signature. Goroutines within the same group have identical
// (or similar, depending on strategy) call stacks.
type Group struct {
	Signature  string              // normalized stack signature used for grouping
	State      string              // most common goroutine state in this group
	Goroutines []parser.Goroutine  // all goroutines in this group
}

// Count returns the number of goroutines in the group.
func (g *Group) Count() int {
	return len(g.Goroutines)
}

// TopFunction returns the top-level function name from the first
// goroutine's stack trace. This is the function at the top of the
// call stack (where the goroutine is currently executing/blocked).
func (g *Group) TopFunction() string {
	if len(g.Goroutines) == 0 || len(g.Goroutines[0].Frames) == 0 {
		return "<unknown>"
	}
	return g.Goroutines[0].Frames[0].Function
}

// TopFile returns the file:line from the top stack frame.
func (g *Group) TopFile() string {
	if len(g.Goroutines) == 0 || len(g.Goroutines[0].Frames) == 0 {
		return "<unknown>"
	}
	f := g.Goroutines[0].Frames[0]
	return f.File
}

// Summary returns a one-line human-readable summary of the group.
func (g *Group) Summary() string {
	return g.TopFunction() + " (" + g.State + ")"
}
