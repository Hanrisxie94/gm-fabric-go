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

package proxymeta

import (
	"fmt"
	"net/http"

	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
)

// MetaOption returns a ServeMuxOption that captures headers we are interested
// in and returns gRPC metadata
func MetaOption() runtime.ServeMuxOption {

	f := func(ctx oldcontext.Context, req *http.Request) metadata.MD {
		metaMap := make(map[string]string)
		for _, hkey := range headers.HeadersOfInterest {
			if hvalue := req.Header.Get(hkey); hvalue != "" {
				metaMap[hkey] = hvalue
			}
		}

		// stuff the HTTP route into the gRPC meta so we can track it
		metaMap[headers.PrevRouteHeader] =
			fmt.Sprintf("%s/%s", req.URL.EscapedPath(), req.Method)

		return metadata.New(metaMap)
	}

	return runtime.WithMetadata(f)
}
