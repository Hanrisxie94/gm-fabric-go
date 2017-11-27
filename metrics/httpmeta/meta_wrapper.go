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

package httpmeta

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/metadata"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
)

// wrappedMeta wraps a single Handler
type wrappedMeta struct {
	next http.Handler
}

// HandlerFunc returns an http.HandlerFunc
func HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return wrappedMeta{next: next}.ServeHTTP
}

// Handler returns an http.Handler
func Handler(next http.Handler) http.Handler {
	return wrappedMeta{next: next}
}

// ServeHTTP implements the http.Handler interface
// It updates grpc metadata int the request Context with selected HTTP headers
// This can serve as a HandlerFunc
func (wm wrappedMeta) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// FYI: The B3 portion of the header is so named for the original name of
	// Zipkin: BigBrotherBird.

	metaMap := make(map[string]string)
	for _, hkey := range headers.HeadersOfInterest {
		if hvalue := req.Header.Get(hkey); hvalue != "" {
			metaMap[hkey] = hvalue
		}
	}

	// stuff the HTTP route into the gRPC meta so we can track it
	metaMap[headers.PrevRouteHeader] =
		fmt.Sprintf("%s/%s", req.URL.EscapedPath(), req.Method)

	inMd, _ := metadata.FromIncomingContext(req.Context())
	outMd := metadata.Join(inMd, metadata.New(metaMap))
	outCtx := metadata.NewOutgoingContext(req.Context(), outMd)
	outReq := req.WithContext(outCtx)

	wm.next.ServeHTTP(w, outReq)
}
