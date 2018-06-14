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
	"all/latency_ms.avg": 3196.968974,
	"all/latency_ms.count": 419,
	"all/latency_ms.max": 5389,
	"all/latency_ms.min": 637,
	"all/latency_ms.sum": 1339530,
	"all/latency_ms.p50": 3261,
	"all/latency_ms.p90": 4415,
	"all/latency_ms.p95": 4722,
	"all/latency_ms.p99": 5067,
	"all/latency_ms.p9990": 5389,
	"all/latency_ms.p9999": 5389,
	"all/errors.count": 0,
	"all/in_throughput": 43254,
	"all/out_throughput": 47971,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.avg": 2973.338095,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.count": 210,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.max": 5171,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.min": 867,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.sum": 624401,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p50": 2953,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p90": 4062,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p95": 4534,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p99": 5025,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p9990": 5171,
	"/metricstester.MetricsTester/CatalogStream/latency_ms.p9999": 5171,
	"/metricstester.MetricsTester/CatalogStream/errors.count": 0,
	"/metricstester.MetricsTester/CatalogStream/in_throughput": 21414,
	"/metricstester.MetricsTester/CatalogStream/out_throughput": 45672,
	"/metricstester.MetricsTester/OrderItem/latency_ms.avg": 3421.669856,
	"/metricstester.MetricsTester/OrderItem/latency_ms.count": 209,
	"/metricstester.MetricsTester/OrderItem/latency_ms.max": 5389,
	"/metricstester.MetricsTester/OrderItem/latency_ms.min": 637,
	"/metricstester.MetricsTester/OrderItem/latency_ms.sum": 715129,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p50": 3570,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p90": 4610,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p95": 4825,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p99": 5067,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p9990": 5389,
	"/metricstester.MetricsTester/OrderItem/latency_ms.p9999": 5389,
	"/metricstester.MetricsTester/OrderItem/errors.count": 0,
	"/metricstester.MetricsTester/OrderItem/in_throughput": 21840,
	"/metricstester.MetricsTester/OrderItem/out_throughput": 2299
}
 ```

 go-metrics
 ```JSON    
 $ curl 127.0.0.1:20001/gometrics
{
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/heap_objects": 15482.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/total_gc_pause_ns": 2753584.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/total_gc_runs": 22.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/num_goroutines": 10.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/alloc_bytes": 1945464.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/sys_bytes": 8427768.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/malloc_count": 277101.000000,
	"(Gauge)/test-http-server/c7c1e60c6d8b/runtime/free_count": 261619.000000,
	"(EmitKey)/test-http-server/httpserver/GET": 0.000000,
	"(EmitKey)/test-http-server/httpserver/POST": 0.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1002": 54.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1008": 39.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1003": 57.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1005": 45.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1006": 55.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1001": 52.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1004": 46.000000,
	"(IncrCounter)/test-http-server/httpserver/sendCatalog": 410.000000,
	"(IncrCounter)/test-http-server/httpserver/orderItem/1007": 61.000000,
	"(AddSample)/test-http-server/httpserver/GetCatalog": 4463.704102,
	"(AddSample)/test-http-server/httpserver/OrderItem": 3199.076172,
	"(AddSample)/test-http-server/runtime/gc_pause_ns": 107059.000000
}
 ```
