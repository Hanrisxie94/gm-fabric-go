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

package prometheus

import (
	"fmt"
	"strings"
	"sync"

	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/metrics/headers"
)

// StatsHandler implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
type StatsHandler struct {
	sync.Mutex
	Collector Collector
	Logger    zerolog.Logger
	StatsData map[string]CollectorEntry
}

// GRPCLoggerOption returns a StatsHandler option function that sets the logger
func GRPCLoggerOption(logger zerolog.Logger) func(*StatsHandler) {
	return func(s *StatsHandler) {
		s.Logger = logger
	}
}

// NewStatsHandler returns an object that implements the stats.Handler interface
// https://godoc.org/google.golang.org/grpc/stats#Handler
func NewStatsHandler(options ...func(*StatsHandler)) (*StatsHandler, error) {
	var s StatsHandler
	var err error

	s.StatsData = make(map[string]CollectorEntry)
	if s.Collector, err = NewCollector(); err != nil {
		return nil, errors.Wrap(err, "NewCollector")
	}

	for _, f := range options {
		f(&s)
	}

	return &s, nil
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
	_ *stats.ConnTagInfo,
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
		id := inMD[headers.RequestIDHeader]
		if len(id) == 1 {
			requestID = id[0]
		}
		pr := inMD[headers.PrevRouteHeader]
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
	requestID := headers.GetRequestID(ctx)
	if len(requestID) == 0 {
		h.Logger.Error().Str("method", "HandleRPC").
			Msg("Unable to get requestID from context: no metrics")
		return
	}

	h.Lock()
	defer h.Unlock()
	statsEntry := h.StatsData[requestID]
	var statsEnd bool

	switch st := s.(type) {
	case *stats.InHeader:
		// TODO: check for TLS
		statsEntry.Method = "gRPC"
		statsEntry.Key = constructKey(st.FullMethod)
		statsEntry.BytesRead += uint64(st.WireLength)
	case *stats.Begin:
		statsEntry.StartTime = st.BeginTime
	case *stats.InPayload:
		statsEntry.BytesRead += uint64(st.WireLength)
	case *stats.InTrailer:
		statsEntry.BytesRead += uint64(st.WireLength)
	case *stats.OutPayload:
		statsEntry.BytesWritten += uint64(st.WireLength)
	case *stats.OutTrailer:
		statsEntry.BytesWritten += uint64(st.WireLength)
	case *stats.End:
		statsEntry.EndTime = st.EndTime
		statsEnd = true
	}

	if statsEnd {
		if err := h.Collector.Collect(statsEntry); err != nil {
			h.Logger.Error().Err(err).Str("method", "HandleRPC").Msg("Collect")
		}
		delete(h.StatsData, requestID)
	} else {
		h.StatsData[requestID] = statsEntry
	}

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
