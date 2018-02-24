package httputil

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

	"github.com/rs/zerolog"
)

// grpcArrayHandler wraps an HTTP handler to intercept writes to a grpc array
type grpcArrayHandler struct {
	logger zerolog.Logger
	next   http.Handler
}

// HandlerFunc returns an http.HandlerFunc
func HandlerFunc(logger zerolog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return grpcArrayHandler{logger: logger, next: next}.ServeHTTP
}

// Handler returns an http.Handler
func Handler(logger zerolog.Logger, next http.Handler) http.Handler {
	return grpcArrayHandler{logger: logger, next: next}
}

// ServeHTTP implements the http.Handler interface
func (gr grpcArrayHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw := newGRPCArrayWriter(gr.logger, w)
	gr.next.ServeHTTP(gw, req)
	if err := gw.flush(); err != nil {
		// there's not much we can do if we get an error at this point
		gr.logger.Error().AnErr("ServeHTTP", err).Msg("")
	}
}
