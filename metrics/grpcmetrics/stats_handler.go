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
	"fmt"
	"strings"
	"time"

	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// StatsHandler implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
type StatsHandler struct {
	metricsChan chan<- subject.MetricsEvent
	tags        []string
}

// NewStatsHandler returns an object that implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
func NewStatsHandler(metricsChan chan<- subject.MetricsEvent) *StatsHandler {
	var h StatsHandler
	h.metricsChan = metricsChan

	return &h
}

// NewStatsHandlerWithTags returns an object that implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
// The tags will be added to each MetricsEvent
func NewStatsHandlerWithTags(
	metricsChan chan<- subject.MetricsEvent,
	tags []string,
) *StatsHandler {
	var h StatsHandler
	h.metricsChan = metricsChan
	h.tags = tags

	return &h
}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//  - On server side, the context used in HandleRPC for all RPCs on this
// connection will be derived from the context returned.
//  - On client side, the context is not derived from the context returned.
func (h *StatsHandler) TagConn(
	ctx oldcontext.Context,
	info *stats.ConnTagInfo,
) oldcontext.Context {
	return ctx
}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from
// the returned context.
func (h *StatsHandler) TagRPC(
	ctx oldcontext.Context,
	info *stats.RPCTagInfo,
) oldcontext.Context {
	var requestID string
	var prevRoute string

	if inMD, ok := metadata.FromIncomingContext(ctx); ok {
		id, _ := inMD[headers.RequestIDHeader]
		if len(id) == 1 {
			requestID = id[0]
		}
		pr, _ := inMD[headers.PrevRouteHeader]
		if len(pr) == 1 {
			prevRoute = pr[0]
		}
	}

	if requestID == "" {
		requestID = headers.NewRequestID()
	}

	// return a context with the values set
	return headers.SetRequestID(headers.SetPrevRoute(ctx, prevRoute), requestID)
}

// HandleConn processes the Conn stats.
func (h *StatsHandler) HandleConn(
	ctx oldcontext.Context,
	s stats.ConnStats,
) {
}

// HandleRPC processes the RPC stats.
func (h *StatsHandler) HandleRPC(
	ctx oldcontext.Context,
	s stats.RPCStats,
) {
	var event subject.MetricsEvent
	event.Timestamp = time.Now()
	event.RequestID = headers.GetRequestID(ctx)
	event.PrevRoute = headers.GetPrevRoute(ctx)
	event.Tags = h.tags

	switch st := s.(type) {
	case *stats.InHeader:
		event.EventType = "rpc.InHeader"
		// TODO: check for TLS
		event.Transport = subject.EventTransportRPC
		event.Key = constructKey(st.FullMethod)
		event.Value = int64(st.WireLength)
		event.Tags = append(event.Tags, subject.JoinTag("FullMethod", st.FullMethod))
	case *stats.Begin:
		event.EventType = "rpc.Begin"
		event.Timestamp = st.BeginTime
	case *stats.InPayload:
		event.EventType = "rpc.InPayload"
		event.Value = int64(st.WireLength)
	case *stats.InTrailer:
		event.EventType = "rpc.InTrailer"
		event.Value = int64(st.WireLength)
	case *stats.OutPayload:
		event.EventType = "rpc.OutPayload"
		event.Value = int64(st.WireLength)
	case *stats.OutTrailer:
		event.EventType = "rpc.OutTrailer"
		event.Value = int64(st.WireLength)
	case *stats.End:
		event.EventType = "rpc.End"
		event.Timestamp = st.EndTime
		event.Value = st.Error
	}

	h.metricsChan <- event

}

// constructKey takes the full grpc method name, of the form
// '/metricstester.MetricsTester/CatalogStream'
// and returns
// 'function/CatalogStream'
func constructKey(fullMethod string) string {
	var funcName string
	s := strings.Split(fullMethod, "/")
	if len(s) > 0 {
		funcName = s[len(s)-1]
	}
	return fmt.Sprintf("function/%s", funcName)
}
