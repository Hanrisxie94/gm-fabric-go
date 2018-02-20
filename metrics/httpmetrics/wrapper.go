// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpmetrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

var (
	// httpMetrics is a package level instance of HTTPMetrics
	// it should only be used for memoizing New
	httpMetrics *HTTPMetrics
)

// HTPStatusError implements the error interface to report a HTTP
// status
type HTPStatusError struct {
	StatusCode int
}

// Error implements the 'error' interface for an HTTP status error
func (e HTPStatusError) Error() string {
	return fmt.Sprintf("HTTP: %d", e.StatusCode)
}

// HTTPMetrics supports HTTP metrics
// This is a wrapper for the stats chan which feeds the metricsserver
type HTTPMetrics struct {
	metricsChan chan<- subject.MetricsEvent
	tags        []string
	RollupPath  string
}

// wrappedMetrics wraps a single Handler
type wrappedMetrics struct {
	h          *HTTPMetrics
	next       http.Handler
	rollupPath string
}

// countWriter implements the http.ResponseWriter protocol to
// capture soe counts
type countWriter struct {
	status       int
	bytesWritten int64
	next         http.ResponseWriter
}

// New returns an object that supports HTTP metrics
// This basically passes on the stats chan to individal handlers
func New(MetricsChan chan<- subject.MetricsEvent) *HTTPMetrics {
	if httpMetrics == nil {
		httpMetrics = &HTTPMetrics{metricsChan: MetricsChan}
	}
	return httpMetrics
}

// NewWithTags returns an object that supports HTTP metrics
// This basically passes on the stats chan to individal handlers
func NewWithTags(
	metricsChan chan<- subject.MetricsEvent,
	tags []string,
) *HTTPMetrics {
	if httpMetrics == nil {
		httpMetrics = &HTTPMetrics{metricsChan: metricsChan, tags: tags}
	}
	return httpMetrics
}

// HandlerFunc returns an http.HandlerFunc
func (h *HTTPMetrics) HandlerFunc(next http.HandlerFunc, options ...func(*HTTPMetrics)) http.HandlerFunc {
	return h.Handler(next, options...).ServeHTTP
}

// Handler returns an http.Handler
func (h *HTTPMetrics) Handler(next http.Handler, options ...func(*HTTPMetrics)) http.Handler {
	wm := wrappedMetrics{h: h, next: next}
	for _, option := range options {
		option(h)
	}
	// ServeHTTP can only reference h because wm doesn't exist when creating options,
	// so inject rollupPath passed in as an option exported through h
	wm.rollupPath = h.RollupPath
	return wm
}

// ServeHTTP implements the http.Handler interface
// This can server as a HandlerFunc
// Just capture as much information about the transacton as we can
// and pass it on to the stats server
func (wm wrappedMetrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var transport subject.EventTransport
	if req.TLS == nil {
		transport = subject.EventTransportHTTP
	} else {
		transport = subject.EventTransportHTTPS
	}

	requestID := req.Header.Get(headers.RequestIDHeader)
	if requestID == "" {
		requestID = headers.NewRequestID()
		req.Header.Add(headers.RequestIDHeader, requestID)
	}

	path := req.URL.EscapedPath()
	if len(wm.rollupPath) > 0 {
		if strings.HasPrefix(path, wm.rollupPath) {
			path = wm.rollupPath[0 : len(wm.rollupPath)-1]
		}
	}
	wm.h.metricsChan <- subject.MetricsEvent{
		EventType: "rpc.InHeader",
		Transport: transport,
		RequestID: requestID,
		Timestamp: time.Now(),
		Key:       fmt.Sprintf("route%s/%s", path, req.Method),
		Tags:      wm.h.tags,
	}
	wm.h.metricsChan <- subject.MetricsEvent{
		EventType: "rpc.Begin",
		RequestID: requestID,
		Timestamp: time.Now(),
		Tags:      wm.h.tags,
	}
	wm.h.metricsChan <- subject.MetricsEvent{
		EventType: "rpc.InPayload",
		RequestID: requestID,
		Timestamp: time.Now(),
		Value:     req.ContentLength,
		Tags:      wm.h.tags,
	}

	c := countWriter{next: w}
	wm.next.ServeHTTP(&c, req)
	status := c.status
	if status == 0 {
		status = http.StatusOK
	}

	// Note: Ignore codes less than 400.
	// We may want to count some special cases.
	// too noisy to even count: 200, 206, 304

	// A 4xx error should indicate bad input
	if 400 <= c.status && c.status < 500 {
		wm.h.metricsChan <- subject.MetricsEvent{
			EventType:  "rpc.End",
			RequestID:  requestID,
			Timestamp:  time.Now(),
			HTTPStatus: status,
			Tags:       wm.h.tags,
		}
	}

	// A 5xx error definitely must be counted as a problem
	if 500 <= c.status {
		wm.h.metricsChan <- subject.MetricsEvent{
			EventType:  "rpc.End",
			RequestID:  requestID,
			Timestamp:  time.Now(),
			HTTPStatus: status,
			Tags:       wm.h.tags,
		}
	} else {
		wm.h.metricsChan <- subject.MetricsEvent{
			EventType: "rpc.OutPayload",
			RequestID: requestID,
			Timestamp: time.Now(),
			Value:     c.bytesWritten,
			Tags:      wm.h.tags,
		}
		wm.h.metricsChan <- subject.MetricsEvent{
			EventType:  "rpc.End",
			RequestID:  requestID,
			Timestamp:  time.Now(),
			HTTPStatus: status,
			Tags:       wm.h.tags,
		}
	}
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
//
// Changing the header map after a call to WriteHeader (or
// Write) has no effect unless the modified headers are
// trailers.
//
// There are two ways to set Trailers. The preferred way is to
// predeclare in the headers which trailers you will later
// send by setting the "Trailer" header to the names of the
// trailer keys which will come later. In this case, those
// keys of the Header map are treated as if they were
// trailers. See the example. The second way, for trailer
// keys not known to the Handler until after the first Write,
// is to prefix the Header map keys with the TrailerPrefix
// constant value. See TrailerPrefix.
//
// To suppress implicit response headers (such as "Date"), set
// their value to nil.
func (c *countWriter) Header() http.Header {
	return c.next.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
//
// If WriteHeader has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data. If the Header
// does not contain a Content-Type line, Write adds a Content-Type set
// to the result of passing the initial 512 bytes of written data to
// DetectContentType.
//
// Depending on the HTTP protocol version and the client, calling
// Write or WriteHeader may prevent future reads on the
// Request.Body. For HTTP/1.x requests, handlers should read any
// needed request body data before writing the response. Once the
// headers have been flushed (due to either an explicit Flusher.Flush
// call or writing enough data to trigger a flush), the request body
// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
// handlers to continue to read the request body while concurrently
// writing the response. However, such behavior may not be supported
// by all HTTP/2 clients. Handlers should read before writing if
// possible to maximize compatibility.
func (c *countWriter) Write(data []byte) (int, error) {
	n, err := c.next.Write(data)
	c.bytesWritten += int64(n)

	return n, err
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (c *countWriter) WriteHeader(status int) {
	c.status = status
	c.next.WriteHeader(status)
}
