# Fabric Prometheus Metrics

## Capturing

### HTTP metrics

* create a ```Collector``` for Prometheus metrics
* compose an HTTP handler into the handlers for each method

```go
import (
    pm "github.com/deciphernow/gm-fabric-go/metrics/prometheus"
)

    collector, err := pm.NewCollector()
    if err != nil {
        log.Fatalf("pm.NewCollector: %s", err)
    }

    mux := http.NewServeMux()

    catalogHandler :=
        pm.NewHandler(
            collector,
            httpm.HandlerFunc(
                http.HandlerFunc(state.handleCatalog),
            ),
            pm.HTTPLoggerOption(logger),
        )

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

### gRPC Metrics

* create a Prometheus metrics StatsHandler
* add it to server options

Note that gRPC will only report to one StatsHandler: whichever one
gets added last gets the stats.

```go
import (
    pm "github.com/deciphernow/gm-fabric-go/metrics/prometheus"
)

    pmStatsHandler, err := pm.NewStatsHandler(pm.GRPCLoggerOption(logger))
    if err != nil {
        logger.Fatal().Err(err).Msg("pm.NewStatsHandler")
    }

    opts := []grpc.ServerOption{
        grpc.StatsHandler(pmStatsHandler),
    }

    grpcServer := grpc.NewServer(opts...)
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

    metricsMux := http.NewServeMux()
    metricsMux.Handle(
        viper.GetString("metrics_dashboard_prometheus_uri_path"),
        metricsserver.NewDashboardHandler(promReporter.Report, miscReporter.Report),
    )
```

You must use the Prometheus job name to connect the metrics yoou are capturing with
the metrics you report.

```yml
scrape_configs:

  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'store-http'
```
