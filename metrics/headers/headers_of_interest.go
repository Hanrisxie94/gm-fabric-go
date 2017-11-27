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

package headers

import (
	"github.com/google/uuid"
	oldcontext "golang.org/x/net/context"
)

const (
	// RequestIDHeader the header we use to identify requests
	RequestIDHeader = "x-request-id"

	// PrevRouteHeader stores the route leading to a gRPC method
	// (Decipher internal use)
	PrevRouteHeader = "x-gm-prev-route"

	// PrevMethodHeader stores the gRPV method leading to a gRPC method
	// (Decipher internal use)
	PrevMethodHeader = "x-gm-prev-method"
)

// HeadersOfInterest are HTTP Headers (or grpc metadata) that we are interested in.
//
// FYI: b3 is the internal name for Zipkin (big brother bird)
var HeadersOfInterest = []string{
	"x-ot-span-context",
	RequestIDHeader,
	"x-b3-traceid",
	"x-b3-spanid",
	"x-bs-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
	"x-client-trace-id",
	"x-envoy-downstream-service-cluster",
	"x-envoy-downstream-service-node",
	"x-envoy-external-address",
	"x-envoy-force-trace",
	"x-envoy-internal",
	"x-envoy-ip-tags",
	PrevRouteHeader,
	PrevMethodHeader,
}

// RequestIDCtxKey stores request id, either generated or from "x-request-id"
type requestIDCtxKey struct{}

// PrevRouteCtxKey stores  the content of the "x-gm-prev-route" header (if any)
type prevRouteCtxKey struct{}

// NewRequestID creates a request is
func NewRequestID() string {
	return uuid.New().String()
}

// SetRequestID sets our internal value for x-request-id in the context
func SetRequestID(ctx oldcontext.Context, requestID string) oldcontext.Context {
	if ctx == nil {
		ctx = oldcontext.Background()
	}
	return oldcontext.WithValue(ctx, requestIDCtxKey{}, requestID)
}

// GetRequestID retrieves our internal value for x-request-id from the context
func GetRequestID(ctx oldcontext.Context) string {
	if ctx == nil {
		return ""
	}

	value := ctx.Value(requestIDCtxKey{})

	if value == nil {
		return ""
	}

	return value.(string)
}

// SetPrevRoute sets our internal value for x-gm-prev-route in the context
func SetPrevRoute(ctx oldcontext.Context, prevRoute string) oldcontext.Context {
	if ctx == nil {
		ctx = oldcontext.Background()
	}
	return oldcontext.WithValue(ctx, prevRouteCtxKey{}, prevRoute)
}

// GetPrevRoute retrieves our internal value for x-gm-prev-route from the context
func GetPrevRoute(ctx oldcontext.Context) string {
	if ctx == nil {
		return ""
	}

	value := ctx.Value(prevRouteCtxKey{})

	if value == nil {
		return ""
	}

	return value.(string)
}
