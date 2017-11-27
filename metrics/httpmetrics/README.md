# Metrics HTTP Wrapper #

The ```httpmetrics``` package provides a wrapper for HTTP handlers to capture
GM Fabric metrics.

It provides methods to wrap go [HTTP handler functions](https://golang.org/pkg/net/http/#example_ServeMux_Handle)

## Example ##

 * create a standard metrics observer,
 * create an http metrics manager
 * wrap a standard http handler functionson

```go
include(
    "github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
    ms "github.com/deciphernow/gm-fabric-go/metrics/metricsserver"
)

msObserver := ms.New(
    ctx,
    env.MetricsServerAddress(),
    env.MetricsCacheSize(),
    env.MetricsURIPath(),
)

metricsChan := subject.New(ctx, msObserver)

httpm := httpmetrics.New(metricsChan)

http.HandleFunc(apiPath, httpm.HandlerFunc(store.storeHandler))

http.ListenAndServe(serverAddr, nil)

```

We can also wrap HTTP muxers, including gorilla with
