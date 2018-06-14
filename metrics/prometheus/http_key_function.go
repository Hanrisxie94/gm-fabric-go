package prometheus

import "net/http"

// HTTPKeyFunc is a function for defining the 'key' label in HTTP metrics
type HTTPKeyFunc func(*http.Request) string

var (
	// DefaultHTTPKeyFunc returns the URI as the metrics key
	DefaultHTTPKeyFunc HTTPKeyFunc = func(req *http.Request) string {
		return req.URL.EscapedPath()
	}
)
