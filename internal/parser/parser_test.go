package parser

import (
	"os"
	"testing"
)

func TestParseSingleGoroutine(t *testing.T) {
	input := `goroutine 1 [running]:
main.main()
	/app/main.go:15 +0x1a2
runtime.main()
	/usr/local/go/src/runtime/proc.go:267 +0x207
`

	goroutines, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(goroutines) != 1 {
		t.Fatalf("expected 1 goroutine, got %d", len(goroutines))
	}

	g := goroutines[0]
	if g.ID != 1 {
		t.Errorf("expected ID 1, got %d", g.ID)
	}
	if g.State != "running" {
		t.Errorf("expected state 'running', got %q", g.State)
	}
	if len(g.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(g.Frames))
	}

	// First frame (top of stack)
	f := g.Frames[0]
	if f.Function != "main.main" {
		t.Errorf("frame 0: expected function 'main.main', got %q", f.Function)
	}
	if f.File != "/app/main.go" {
		t.Errorf("frame 0: expected file '/app/main.go', got %q", f.File)
	}
	if f.Line != 15 {
		t.Errorf("frame 0: expected line 15, got %d", f.Line)
	}

	// Second frame
	f = g.Frames[1]
	if f.Function != "runtime.main" {
		t.Errorf("frame 1: expected function 'runtime.main', got %q", f.Function)
	}
}

func TestParseMultipleGoroutines(t *testing.T) {
	input := `goroutine 6 [IO wait]:
net/http.(*connReader).Read(0xc0001a6000, {0xc0001c8000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:786 +0x158

goroutine 7 [IO wait]:
net/http.(*connReader).Read(0xc0001a6100, {0xc0001c9000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:786 +0x158

goroutine 9 [chan receive]:
main.worker(0xc000100060)
	/app/worker.go:28 +0x56
`

	goroutines, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(goroutines) != 3 {
		t.Fatalf("expected 3 goroutines, got %d", len(goroutines))
	}

	// Check IDs
	expectedIDs := []int{6, 7, 9}
	for i, expected := range expectedIDs {
		if goroutines[i].ID != expected {
			t.Errorf("goroutine %d: expected ID %d, got %d", i, expected, goroutines[i].ID)
		}
	}

	// Check states
	if goroutines[0].State != "IO wait" {
		t.Errorf("goroutine 0: expected state 'IO wait', got %q", goroutines[0].State)
	}
	if goroutines[2].State != "chan receive" {
		t.Errorf("goroutine 2: expected state 'chan receive', got %q", goroutines[2].State)
	}
}

func TestParseStateDuration(t *testing.T) {
	input := `goroutine 5 [chan receive, 5 minutes]:
main.worker(0xc000100060)
	/app/worker.go:28 +0x56
`

	goroutines, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if goroutines[0].State != "chan receive" {
		t.Errorf("expected state 'chan receive' (duration stripped), got %q", goroutines[0].State)
	}
}

func TestParseCreatedByFrame(t *testing.T) {
	input := `goroutine 9 [chan receive]:
main.worker(0xc000100060)
	/app/worker.go:28 +0x56
created by main.startWorkers in goroutine 1
	/app/main.go:45 +0x85
`

	goroutines, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goroutines[0]
	if len(g.Frames) != 2 {
		t.Fatalf("expected 2 frames (including created-by), got %d", len(g.Frames))
	}

	// The "created by" frame should capture the function name
	f := g.Frames[1]
	if f.Function != "main.startWorkers" {
		t.Errorf("created-by frame: expected function 'main.startWorkers', got %q", f.Function)
	}
	if f.File != "/app/main.go" {
		t.Errorf("created-by frame: expected file '/app/main.go', got %q", f.File)
	}
	if f.Line != 45 {
		t.Errorf("created-by frame: expected line 45, got %d", f.Line)
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParseNoGoroutines(t *testing.T) {
	_, err := Parse("some random text\nthat is not a goroutine dump\n")
	if err == nil {
		t.Fatal("expected error for input with no goroutines, got nil")
	}
}

func TestParseSampleDumpFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample_dump.txt")
	if err != nil {
		t.Fatalf("failed to read sample dump file: %v", err)
	}

	goroutines, err := Parse(string(data))
	if err != nil {
		t.Fatalf("unexpected error parsing sample dump: %v", err)
	}

	// The sample dump has 11 goroutines (IDs 1, 6-15)
	if len(goroutines) != 11 {
		t.Errorf("expected 11 goroutines from sample dump, got %d", len(goroutines))
	}

	// Check that all goroutines have at least one frame
	for _, g := range goroutines {
		if len(g.Frames) == 0 {
			t.Errorf("goroutine %d has no frames", g.ID)
		}
	}

	// Verify specific goroutines
	// Goroutine 1 should be [running]
	if goroutines[0].State != "running" {
		t.Errorf("goroutine 1: expected state 'running', got %q", goroutines[0].State)
	}

	// Goroutines 6, 7, 8 should be [IO wait]
	for i := 1; i <= 3; i++ {
		if goroutines[i].State != "IO wait" {
			t.Errorf("goroutine %d: expected state 'IO wait', got %q", goroutines[i].ID, goroutines[i].State)
		}
	}

	// Goroutines 9-12 should be [chan receive]
	for i := 4; i <= 7; i++ {
		if goroutines[i].State != "chan receive" {
			t.Errorf("goroutine %d: expected state 'chan receive', got %q", goroutines[i].ID, goroutines[i].State)
		}
	}
}
