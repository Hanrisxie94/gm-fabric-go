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

package grpcclient

import (
	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
)

// UnaryClientInterceptor intercepts the execution of a unary RPC on the client.
// It captures some metadata from context, then calls 'invoker' to pass control.
func UnaryClientInterceptor(
	ctx oldcontext.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(adjustContext(ctx), method, req, reply, cc, opts...)
}

// StreamClientInterceptor intercepts the creation of ClientStream.
// It captures some metadata from context, then calls 'streamer' to pass control.
func StreamClientInterceptor(
	ctx oldcontext.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return streamer(adjustContext(ctx), desc, cc, method, opts...)
}

func adjustContext(ctx oldcontext.Context) oldcontext.Context {
	metaMap := make(map[string]string)

	if inMD, ok := metadata.FromIncomingContext(ctx); ok {
		for _, hkey := range headers.HeadersOfInterest {
			if hval, ok := inMD[hkey]; ok {
				metaMap[hkey] = hval[0]
			}
		}
	}

	requestID := headers.GetRequestID(ctx)
	if requestID != "" {
		metaMap[headers.RequestIDHeader] = requestID
	}

	return metadata.NewOutgoingContext(ctx, metadata.New(metaMap))
}
