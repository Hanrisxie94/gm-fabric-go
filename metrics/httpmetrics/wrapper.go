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
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
	"github.com/deciphernow/gm-fabric-go/metrics/keyfunc"
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
	keyFunc     keyfunc.HTTPKeyFunc
}

// wrappedMetrics wraps a single Handler
type wrappedMetrics struct {
	h       *HTTPMetrics
	next    http.Handler
	keyFunc keyfunc.HTTPKeyFunc
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
	// Users need to be able to send metrics to specific counting keys
	wm.keyFunc = h.keyFunc
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

	if wm.keyFunc == nil {
		wm.keyFunc = keyfunc.DefaultHTTPKeyFunc
	}
	key := wm.keyFunc(req)

	wm.h.metricsChan <- subject.MetricsEvent{
		EventType: "rpc.InHeader",
		Transport: transport,
		RequestID: requestID,
		Timestamp: time.Now(),
		Key:       fmt.Sprintf("route%s/%s", key, req.Method),
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

	c := CountWriter{Next: w}
	wm.next.ServeHTTP(&c, req)
	status := c.Status
	if status == 0 {
		status = http.StatusOK
	}

	wm.h.metricsChan <- subject.MetricsEvent{
		EventType: "rpc.OutPayload",
		RequestID: requestID,
		Timestamp: time.Now(),
		Value:     c.BytesWritten,
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
