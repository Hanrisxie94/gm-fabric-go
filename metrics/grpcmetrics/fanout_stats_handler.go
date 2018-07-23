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

package grpcmetrics

import (
	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc/stats"
)

// FanoutHandler implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
type FanoutHandler struct {
	inner []stats.Handler
}

// NewFanoutHandler returns an object that implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
func NewFanoutHandler(handlers ...stats.Handler) FanoutHandler {
	return FanoutHandler{inner: handlers}
}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//  - On server side, the context used in HandleRPC for all RPCs on this
// connection will be derived from the context returned.
//  - On client side, the context is not derived from the context returned.
func (h FanoutHandler) TagConn(
	ctx oldcontext.Context,
	info *stats.ConnTagInfo,
) oldcontext.Context {
	for _, handler := range h.inner {
		ctx = handler.TagConn(ctx, info)
	}
	return ctx
}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from
// the returned context.
func (h FanoutHandler) TagRPC(
	ctx oldcontext.Context,
	info *stats.RPCTagInfo,
) oldcontext.Context {
	for _, handler := range h.inner {
		ctx = handler.TagRPC(ctx, info)
	}
	return ctx
}

// HandleConn processes the Conn stats.
func (h FanoutHandler) HandleConn(
	ctx oldcontext.Context,
	s stats.ConnStats,
) {
	for _, handler := range h.inner {
		handler.HandleConn(ctx, s)
	}
}

// HandleRPC processes the RPC stats.
func (h FanoutHandler) HandleRPC(
	ctx oldcontext.Context,
	s stats.RPCStats,
) {
	for _, handler := range h.inner {
		handler.HandleRPC(ctx, s)
	}
}
