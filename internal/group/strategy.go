package group

import (
	"sort"
	"strings"

	"github.com/santhosh-005/goroviz/internal/parser"
)

// Strategy defines the interface for goroutine grouping algorithms.
// Different strategies can group goroutines using different heuristics
// (exact match, prefix match, similarity clustering, etc.).
type Strategy interface {
	// Group takes a slice of parsed goroutines and returns them
	// organized into groups based on the strategy's algorithm.
	Group(goroutines []parser.Goroutine) []Group
}

// ExactMatchStrategy groups goroutines that have identical stack
// frame signatures. The signature is computed by joining all function
// names in the stack trace.
//
// This is the simplest and most precise grouping strategy — two
// goroutines are in the same group only if their entire call stacks
// match exactly (ignoring memory addresses and arguments).
type ExactMatchStrategy struct{}

// NewExactMatchStrategy creates a new ExactMatchStrategy.
func NewExactMatchStrategy() *ExactMatchStrategy {
	return &ExactMatchStrategy{}
}

// Group implements Strategy. It groups goroutines by exact stack
// signature match, then sorts groups by count (largest first).
func (s *ExactMatchStrategy) Group(goroutines []parser.Goroutine) []Group {
	if len(goroutines) == 0 {
		return nil
	}

	// Build signature → goroutines map
	groupMap := make(map[string][]parser.Goroutine)
	stateMap := make(map[string]map[string]int) // signature → state → count

	for _, g := range goroutines {
		sig := signature(g)
		groupMap[sig] = append(groupMap[sig], g)

		if stateMap[sig] == nil {
			stateMap[sig] = make(map[string]int)
		}
		stateMap[sig][g.State]++
	}

	// Convert map to slice of Groups
	groups := make([]Group, 0, len(groupMap))
	for sig, grs := range groupMap {
		groups = append(groups, Group{
			Signature:  sig,
			State:      dominantState(stateMap[sig]),
			Goroutines: grs,
		})
	}

	// Sort by count descending (largest groups first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count() > groups[j].Count()
	})

	return groups
}

// signature generates a normalized stack signature for a goroutine
// by joining all function names with " -> ".
// Example: "net/http.(*connReader).Read -> bufio.(*Reader).fill -> ..."
func signature(g parser.Goroutine) string {
	if len(g.Frames) == 0 {
		return "<empty>"
	}

	names := make([]string, len(g.Frames))
	for i, f := range g.Frames {
		names[i] = f.Function
	}
	return strings.Join(names, " -> ")
}

// dominantState returns the most frequent state from a state→count map.
func dominantState(states map[string]int) string {
	var maxState string
	var maxCount int
	for state, count := range states {
		if count > maxCount {
			maxCount = count
			maxState = state
		}
	}
	return maxState
}
