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
	"fmt"
	"testing"

	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
)

// invoker implements grpc.UnaryInvoker
// It returns an error if it doesn find RequestIDHeader and PrevRouteHeader
func invoker(
	ctx oldcontext.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	opts ...grpc.CallOption,
) error {
	if headers.GetRequestID(ctx) == "" {
		return fmt.Errorf("request id not found")
	}
	if headers.GetPrevRoute(ctx) == "" {
		return fmt.Errorf("prev route not found")
	}
	return nil
}

func TestUnaryClientInterceptor(t *testing.T) {
	ctx := oldcontext.Background()
	ctx = headers.SetRequestID(ctx, "aaa")
	ctx = headers.SetPrevRoute(ctx, "bbb")

	err := UnaryClientInterceptor(
		ctx,
		"",
		nil,
		nil,
		nil,
		invoker,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
}

// streamer implements grpc.Streamer
// It returns an error if it doesn find RequestIDHeader and PrevRouteHeader
func streamer(
	ctx oldcontext.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	if headers.GetRequestID(ctx) == "" {
		return nil, fmt.Errorf("request id not found")
	}
	if headers.GetPrevRoute(ctx) == "" {
		return nil, fmt.Errorf("prev route not found")
	}
	return nil, nil
}

func TestStreamClientInterceptor(t *testing.T) {
	ctx := oldcontext.Background()
	ctx = headers.SetRequestID(ctx, "aaa")
	ctx = headers.SetPrevRoute(ctx, "bbb")

	_, err := StreamClientInterceptor(
		ctx,
		nil,
		nil,
		"",
		streamer,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
}
