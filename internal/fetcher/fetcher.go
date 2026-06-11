// Package fetcher provides HTTP-based goroutine dump fetching
// from a running Go application's pprof endpoint.
package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// pprofPath is the standard pprof goroutine dump endpoint path.
	pprofPath = "/debug/pprof/goroutine"

	// defaultTimeout for HTTP requests.
	defaultTimeout = 10 * time.Second
)

// Fetch retrieves a goroutine dump from a running Go application's
// pprof endpoint.
//
// The baseURL can be in any of these formats:
//
//	http://localhost:6060
//	http://localhost:6060/
//	http://localhost:6060/debug/pprof/goroutine
//	http://localhost:6060/debug/pprof/goroutine?debug=2
//
// If the URL doesn't already include the pprof path, it is appended
// automatically along with ?debug=2.
func Fetch(baseURL string) (string, error) {
	url := buildURL(baseURL)

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return "", fmt.Errorf("connection refused at %s — is the app running with pprof enabled?\n\nMake sure your app imports:\n  import _ \"net/http/pprof\"", baseURL)
		}
		if strings.Contains(err.Error(), "Timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("timeout connecting to %s — the app may be unresponsive", baseURL)
		}
		return "", fmt.Errorf("failed to connect to %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from %s — is pprof enabled?\n\nMake sure your app imports:\n  import _ \"net/http/pprof\"", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if len(strings.TrimSpace(string(body))) == 0 {
		return "", fmt.Errorf("empty response from %s — no goroutine data returned", url)
	}

	return string(body), nil
}

// buildURL constructs the full pprof goroutine dump URL from a base URL.
func buildURL(baseURL string) string {
	// Ensure scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	// If URL already contains the pprof path, ensure debug=2
	if strings.Contains(baseURL, "/debug/pprof/goroutine") {
		if !strings.Contains(baseURL, "debug=2") {
			if strings.Contains(baseURL, "?") {
				return baseURL + "&debug=2"
			}
			return baseURL + "?debug=2"
		}
		return baseURL
	}

	// Strip trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	return baseURL + pprofPath + "?debug=2"
}
