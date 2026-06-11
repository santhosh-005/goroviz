package group

import (
	"os"
	"testing"

	"github.com/santhosh-005/goroviz/internal/parser"
)

func TestExactMatchGroupIdentical(t *testing.T) {
	// Two goroutines with identical stacks should be in the same group
	goroutines := []parser.Goroutine{
		{
			ID:    1,
			State: "chan receive",
			Frames: []parser.Frame{
				{Function: "main.worker", File: "/app/worker.go", Line: 28},
				{Function: "main.startWorkers", File: "/app/main.go", Line: 45},
			},
		},
		{
			ID:    2,
			State: "chan receive",
			Frames: []parser.Frame{
				{Function: "main.worker", File: "/app/worker.go", Line: 28},
				{Function: "main.startWorkers", File: "/app/main.go", Line: 45},
			},
		},
	}

	strategy := NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	if groups[0].Count() != 2 {
		t.Errorf("expected group count 2, got %d", groups[0].Count())
	}

	if groups[0].State != "chan receive" {
		t.Errorf("expected state 'chan receive', got %q", groups[0].State)
	}
}

func TestExactMatchGroupDifferent(t *testing.T) {
	// Three goroutines: two identical + one different = 2 groups
	goroutines := []parser.Goroutine{
		{
			ID:    1,
			State: "IO wait",
			Frames: []parser.Frame{
				{Function: "net/http.(*connReader).Read", File: "/go/net/http/server.go", Line: 786},
			},
		},
		{
			ID:    2,
			State: "IO wait",
			Frames: []parser.Frame{
				{Function: "net/http.(*connReader).Read", File: "/go/net/http/server.go", Line: 786},
			},
		},
		{
			ID:    3,
			State: "running",
			Frames: []parser.Frame{
				{Function: "main.main", File: "/app/main.go", Line: 15},
			},
		},
	}

	strategy := NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// Largest group should be first (2 goroutines)
	if groups[0].Count() != 2 {
		t.Errorf("expected first group count 2, got %d", groups[0].Count())
	}

	if groups[1].Count() != 1 {
		t.Errorf("expected second group count 1, got %d", groups[1].Count())
	}
}

func TestExactMatchSortByCount(t *testing.T) {
	goroutines := []parser.Goroutine{
		{ID: 1, State: "running", Frames: []parser.Frame{{Function: "a"}}},
		{ID: 2, State: "IO wait", Frames: []parser.Frame{{Function: "b"}}},
		{ID: 3, State: "IO wait", Frames: []parser.Frame{{Function: "b"}}},
		{ID: 4, State: "IO wait", Frames: []parser.Frame{{Function: "b"}}},
		{ID: 5, State: "chan receive", Frames: []parser.Frame{{Function: "c"}}},
		{ID: 6, State: "chan receive", Frames: []parser.Frame{{Function: "c"}}},
	}

	strategy := NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	// Should be sorted: 3, 2, 1
	expectedCounts := []int{3, 2, 1}
	for i, expected := range expectedCounts {
		if groups[i].Count() != expected {
			t.Errorf("group %d: expected count %d, got %d", i, expected, groups[i].Count())
		}
	}
}

func TestExactMatchSingleGoroutine(t *testing.T) {
	goroutines := []parser.Goroutine{
		{
			ID:    42,
			State: "running",
			Frames: []parser.Frame{
				{Function: "main.main", File: "/app/main.go", Line: 10},
			},
		},
	}

	strategy := NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	if groups[0].Count() != 1 {
		t.Errorf("expected count 1, got %d", groups[0].Count())
	}

	if groups[0].TopFunction() != "main.main" {
		t.Errorf("expected top function 'main.main', got %q", groups[0].TopFunction())
	}
}

func TestExactMatchEmptyInput(t *testing.T) {
	strategy := NewExactMatchStrategy()
	groups := strategy.Group(nil)

	if groups != nil {
		t.Errorf("expected nil for empty input, got %v", groups)
	}
}

func TestGroupHelpers(t *testing.T) {
	g := Group{
		Signature: "main.worker -> main.startWorkers",
		State:     "chan receive",
		Goroutines: []parser.Goroutine{
			{
				ID:    1,
				State: "chan receive",
				Frames: []parser.Frame{
					{Function: "main.worker", File: "/app/worker.go", Line: 28},
					{Function: "main.startWorkers", File: "/app/main.go", Line: 45},
				},
			},
		},
	}

	if g.Count() != 1 {
		t.Errorf("Count: expected 1, got %d", g.Count())
	}

	if g.TopFunction() != "main.worker" {
		t.Errorf("TopFunction: expected 'main.worker', got %q", g.TopFunction())
	}

	if g.TopFile() != "/app/worker.go" {
		t.Errorf("TopFile: expected '/app/worker.go', got %q", g.TopFile())
	}

	if g.Summary() != "main.worker (chan receive)" {
		t.Errorf("Summary: expected 'main.worker (chan receive)', got %q", g.Summary())
	}
}

func TestExactMatchWithSampleDump(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample_dump.txt")
	if err != nil {
		t.Fatalf("failed to read sample dump: %v", err)
	}

	goroutines, err := parser.Parse(string(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	strategy := NewExactMatchStrategy()
	groups := strategy.Group(goroutines)

	// The sample dump has:
	// - 3 identical HTTP conn readers (goroutines 6, 7, 8) → 1 group of 3
	// - 4 identical workers (goroutines 9, 10, 11, 12) → 1 group of 4
	// - 2 identical DB connection openers (goroutines 13, 14) → 1 group of 2
	// - 1 main goroutine (goroutine 1) → 1 group of 1
	// - 1 waitgroup goroutine (goroutine 15) → 1 group of 1
	// Total: 5 groups

	if len(groups) != 5 {
		t.Fatalf("expected 5 groups from sample dump, got %d", len(groups))
		for i, g := range groups {
			t.Logf("  group %d: %s (count=%d)", i, g.TopFunction(), g.Count())
		}
	}

	// First group should be the largest (4 workers)
	if groups[0].Count() != 4 {
		t.Errorf("expected largest group to have 4 goroutines, got %d", groups[0].Count())
	}

	// Second group should be 3 HTTP conn readers
	if groups[1].Count() != 3 {
		t.Errorf("expected second group to have 3 goroutines, got %d", groups[1].Count())
	}

	// Third group should be 2 DB openers
	if groups[2].Count() != 2 {
		t.Errorf("expected third group to have 2 goroutines, got %d", groups[2].Count())
	}

	// Verify total goroutine count across all groups
	total := 0
	for _, g := range groups {
		total += g.Count()
	}
	if total != 11 {
		t.Errorf("expected total goroutines across groups = 11, got %d", total)
	}
}
