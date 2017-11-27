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

/*Package metrics contains GM Fabric Metrics Instrumentation

Components

    [Source]     [Source]     [Source]
        |           |            |
        --------------------------
                    |
                [Subject]
                    |
        --------------------------
        |           |            |
    [Observer]  [Observer]   [Observer]
        |           |            |
        --------------------------
                    |
            [Metrics Server]
                    |
        --------------------------
        |           |            |
    [Reporter]  [Reporter]   [Reporter]

Sources

Sources generate metrics data and send MetricsEvent notifications through
the metrics event channel
    * gmfabricsink a go-metrics MetricSink https://github.com/armon/go-metrics#sinks that converts events into MetricsEvent
    * grpcmetrics implements the gRpc stats.Handler https://godoc.org/google.golang.org/grpc/stats#Handler interface to report gRpc events as MetricsEvent
    * httpmetrics implements the go http.Handler](https://golang.org/pkg/net/http/#Handler interface to capture MetricsEvent by wrapping an http.Handler

Subject
The subject is from the observer pattern https://en.wikipedia.org/wiki/Observer_pattern. It maintains a list of observers and notifies each one of incoming MetricsEvent

Observer
Each Observer:
     * consumes a stream of MetricsEvent
     * stores some accumulated data
     * can be a Reporter to make the data available through the Metrics Server

observers:
    * gometricsobserver stores and reports go-metrics events
    * grpcobserver stores and reports gRpc and HTTP events
    * logobserver dumps events to the log
    * sinkobserver feeds events to a go-metrics MetricSink
        * We are particularly interested in using the go-metrics StatsiteSink
        which passes our metrics to a statsd server. See the sample Setup below.

Metrics Server

This is a simply HTTP/TLS sever that will return JSON files summarizing the
observed metrics.
The server will create a single JSON stream from the output of Reporters

Program Setup

The metrics components must be set up in a specific order to satisfy dependencies.

    * Create Observer(s)
    * Create Subject from Observer(s) (This creates the metrics event channel)
    * Start the Metrics Server
    * Create Source(s) tied to the metrics event channel

code sample:
    import gometrics "github.com/armon/go-metrics"

    // Create Observers
    grpcObserver := grpcobserver.New(env.MetricsCacheSize())
    goMetObserver := gometricsobserver.New()

    var statsdSink *gometrics.StatsiteSink
    statsdAddress := env.StatsdAddress() // get statsd address from environment

    statsdSink, err = gometrics.NewStatsiteSink(statsdAddress)
    if err != nil {
        log.Fatalf("gometrics.NewStatsiteSink(%s) failed: %v",
            statsdAddress, err,
        )
    }
    memReportingInterval = time.Minute
    sinkObserver := sinkobserver.New(statsdSink, memReportingInterval)

    // Create Subject from Observer(s) (This creates the metrics event channel)
    metricsChan := subject.New(ctx, grpcObserver, goMetObserver, sinkObserver)

    // Start Metrics Server
    metricServer, err := ms.Start(
        env.MetricsServerAddress(),
        nil, //  *tls.Config
        grpcObserver.Report,
		goMetObserver.Report,
    )
    // handle error ...

    // Create Source(s) tied to the metrics event channel

    goMetricsSource := gmfabricsink.New(metricsChan)
	gometrics.NewGlobal(gometrics.DefaultConfig(serviceName), goMetricsSource)

    // gRpc source
    statsHandler := grpcmetrics.NewStatsHandler(metricsChan)

	opts := []grpc.ServerOption{
			grpc.StatsHandler(statsHandler),
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMetricsTesterServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

Metrics Server Output

    {
    	"all/latency_ms.avg": 3196.968974,
    	"all/latency_ms.count": 419,
    	"all/latency_ms.max": 5389,
    	"all/latency_ms.min": 637,
    	"all/latency_ms.sum": 1339530,
*/
package metrics
