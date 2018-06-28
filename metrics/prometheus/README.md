# Fabric Prometheus Metrics

## Capturing

To capture HTTP metrics, use the HandlerFactory to compose an HTTP handler into the handlers for each method

```go
import (
    pm "github.com/deciphernow/gm-fabric-go/metrics/prometheus"
)

    hf, err := pm.NewSummaryHandlerFactory()
    if err != nil {
        log.Fatalf("pm.NewSummaryHandlerFactory failed: %s", err)
    }

    mux := http.NewServeMux()

    catalogHandler, err := hf.NewHandler(http.HandlerFunc(state.handleCatalog))
    if err != nil {
        log.Fatalf("hf.NewHandler failed: %s", err)
    }

    mux.HandleFunc(catalogPath, catalogHandler.ServeHTTP)
```

Note that the metrics server (or some HTTP server) must make the data available for scraping.

```go
import (
    "githubcom/prometheus/client_golang/prometheus/promhttp"
)

    metricsMux := http.NewServeMux()
    metricsMux.Handle(
        viper.GetString("metrics_prometheus_uri_path"),
        promhttp.Handler(),
    )
```

## Reporting

To report Prometheus metrics to the dashboard you must add the Prometheus reporter to the
metrics server

```go
import(
    "github.com/deciphernow/gm-fabric-go/metrics/metricsserver"
    pm "github.com/deciphernow/gm-fabric-go/metrics/prometheus"
)
    promReporter := pm.PromReporter{
        PrometheusURI: fmt.Sprintf("http://%s", viper.GetString("prometheus_address")),
        JobName:       "store-http", // from prometheus.yml
    }
    miscReporter := metricsserver.NewMiscReporter()
```

You must use the Prometeus job name to connect the metrics yoou are capturing with
the metrics you report.

```yml
scrape_configs:

  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'store-http'
```
