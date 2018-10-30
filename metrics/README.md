# Metrics Instrumentation
## Components
```
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
```
### Sources
Sources generate metrics data and send ```MetricsEvent``` notifications through
the metrics event chanel
 * ```gmfabricsink``` a go-metrics [MetricSink](https://github.com/armon/go-metrics#sinks) that converts events into ```MetricsEvent```
 * ```grpcmetrics``` implements the gRpc [stats.Handler](https://godoc.org/google.golang.org/grpc/stats#Handler) interface to report gRpc events as ```MetricsEvent```
 * ```httpmetrics``` implements the 'go' [http.Handler](https://golang.org/pkg/net/http/#Handler) interface to capture ```MetricsEvent``` by wrapping an ```http.Handler```

### Subject
The subject is from the [observer pattern](https://en.wikipedia.org/wiki/Observer_pattern). It maintains a list of observers and notifies each one of incoming ```MetricsEvent```

### Observer
Each Observer:
 * consumes a stream of ```MetricsEvent```
 * stores some accumulated data
 * can be a Reporter to make the data available through the Metrics Server

observers:
 * ```gometricsobserver``` stores and reports 'go-metrics' events
 * ```grpcobserver``` stores and reports gRpc and HTTP events  
 * ```logobserver``` dumps events to the log
 * ```sinkobserver``` feeds events to a 'go-metrics' ```MetricSink```
     * We are particularly interested in using the go-metrics ```StatsiteSink```
     which passes our metrics to a ```statsd``` server. See the sample Setup below.

### Metrics Server
This is a simply HTTP/TLS sever that will return JSON files summarizing the
observed metrics.

The server will create a single JSON stream from the output of Reporters

## Setup
The metrics components must be set up in a specific order to satisfy dependencies.

 * Create Observer(s)
 * Create Subject from Observer(s) (This creates the metrics event channel)
 * Start the Metrics Server
 * Create Source(s) tied to the metrics event channel


 ```golang
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
 ```  
 ## Metrics Server Output
 gm-fabric-go
 ```JSON
 $ curl 127.0.0.1:10001/metrics
{
	"grey-matter-metrics-version": "1.0.0",
	"Total/requests": 8014,
	"HTTP/requests": 8014,
	"HTTPS/requests": 0,
	"RPC/requests": 0,
	"RPC_TLS/requests": 0,
	"route/acme/services/catalog/GET/requests": 4007,
	"route/acme/services/catalog/GET/routes": "",
	"route/acme/services/catalog/GET/status/200": 4007,
	"route/acme/services/catalog/GET/status/2XX": 4007,
	"route/acme/services/catalog/GET/latency_ms.avg": 1264.595703,
	"route/acme/services/catalog/GET/latency_ms.count": 512,
	"route/acme/services/catalog/GET/latency_ms.max": 2215,
	"route/acme/services/catalog/GET/latency_ms.min": 516,
	"route/acme/services/catalog/GET/latency_ms.sum": 647473,
	"route/acme/services/catalog/GET/latency_ms.p50": 1272,
	"route/acme/services/catalog/GET/latency_ms.p90": 1862,
	"route/acme/services/catalog/GET/latency_ms.p95": 1937,
	"route/acme/services/catalog/GET/latency_ms.p99": 1983,
	"route/acme/services/catalog/GET/latency_ms.p9990": 2215,
	"route/acme/services/catalog/GET/latency_ms.p9999": 2215,
	"route/acme/services/catalog/GET/errors.count": 0,
	"route/acme/services/catalog/GET/in_throughput": 0,
    "route/acme/services/catalog/GET/out_throughput": 299008,
}
 ```

### Computation

#### service gRPC

See: https://godoc.org/google.golang.org/grpc/stats

* Latency = time at rpc.End - time at rpc.Begin
* In-throughput =  wirelength at rpc.InHeader + rpc.InPayload + rpc.InTrailer
* Out-Throughput = wirelength at rpc.OutPayload + rpc.OutTrailer
 
#### service HTTP

See: https://golang.org/pkg/net/http/#Handler

* Latency = time after ServeHTTP - time before
* In-Throughput = size of request body
* Out-throughput = size of response body

#### gm-proxy HTTP

See: https://github.com/DecipherNow/gm-proxy/blob/master/docs/CreatingYourFirstFilter.md

* Latency =  time at OnDestroy - time at DecodeHeaders
* In-Throughput = bytes passed through DecodeData
* Out-throughput = bytes passed through EncodeData
