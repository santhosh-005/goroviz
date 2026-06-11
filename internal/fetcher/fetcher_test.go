package fetcher

import "testing"

func TestBuildURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://localhost:6060",
			expected: "http://localhost:6060/debug/pprof/goroutine?debug=2",
		},
		{
			input:    "http://localhost:6060/",
			expected: "http://localhost:6060/debug/pprof/goroutine?debug=2",
		},
		{
			input:    "localhost:6060",
			expected: "http://localhost:6060/debug/pprof/goroutine?debug=2",
		},
		{
			input:    "http://localhost:6060/debug/pprof/goroutine",
			expected: "http://localhost:6060/debug/pprof/goroutine?debug=2",
		},
		{
			input:    "http://localhost:6060/debug/pprof/goroutine?debug=2",
			expected: "http://localhost:6060/debug/pprof/goroutine?debug=2",
		},
		{
			input:    "https://myapp.example.com:8080",
			expected: "https://myapp.example.com:8080/debug/pprof/goroutine?debug=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := buildURL(tt.input)
			if got != tt.expected {
				t.Errorf("buildURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
