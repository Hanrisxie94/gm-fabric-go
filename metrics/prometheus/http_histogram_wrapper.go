package prometheus

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

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
)

// HandlerState implments the httpHandler
type HandlerState struct {
	collector Collector
	keyFunc   HTTPKeyFunc
	logger    zerolog.Logger
	inner     http.Handler
}

// NewHandler creates a new http.HandleFunc composed with an inntr handler
func NewHandler(
	collector Collector,
	inner http.Handler,
	options ...func(*HandlerState),
) *HandlerState {
	handler := HandlerState{
		collector: collector,
		keyFunc:   DefaultHTTPKeyFunc,
		inner:     inner,
	}

	for _, f := range options {
		f(&handler)
	}

	return &handler
}

// KeyFuncOption returns a function that sets the key function
func KeyFuncOption(keyFunc HTTPKeyFunc) func(*HandlerState) {
	return func(s *HandlerState) {
		s.keyFunc = keyFunc
	}
}

// HTTPLoggerOption returns an options function that sets the loggger
func HTTPLoggerOption(logger zerolog.Logger) func(*HandlerState) {
	return func(s *HandlerState) {
		s.logger = logger
	}
}

// ServeHTTP implements the http.Handler interface
// It collects:
//      http_request_duration_seconds
//      http_request_size_bytes
//      http_response_size_bytes
func (hState *HandlerState) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var entry CollectorEntry

	responseWriter := httpmetrics.CountWriter{Next: w}

	requestReader := httpmetrics.CountReader{Next: req.Body}
	req.Body = &requestReader

	entry.StartTime = time.Now()
	hState.inner.ServeHTTP(&responseWriter, req)
	entry.EndTime = time.Now()

	entry.Key = hState.keyFunc(req)

	entry.Method = normalizeMethod(req.Method)

	entry.Status = normalizeStatus(responseWriter.Status)

	entry.BytesRead = uint64(requestReader.BytesRead)
	entry.BytesWritten = uint64(responseWriter.BytesWritten)

	entry.TLS = req.TLS != nil

	if err := hState.collector.Collect(entry); err != nil {
		hState.logger.Error().Err(err).Msg("Collect")
	}
}

func normalizeMethod(method string) string {
	method = strings.ToUpper(method)
	if method == "" {
		method = "GET"
	}

	return method
}

func normalizeStatus(status int) int {
	if status == 0 {
		status = 200
	}

	return status
}
