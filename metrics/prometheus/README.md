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

Note that gRPC will only report to one StatsHandler. If you want both internal
and Prometheus metrics you must use the fanout handler.

```go
import (
    pm "github.com/deciphernow/gm-fabric-go/metrics/prometheus"
)

	internalStatsHandler := grpcmetrics.NewStatsHandlerWithTags(
		metricsChan,
		statsTags,
    )    
    
    pmStatsHandler, err := pm.NewStatsHandler(pm.GRPCLoggerOption(logger))    
    if err != nil {
        logger.Fatal().Err(err).Msg("pm.NewStatsHandler")
    }

    statsHandler = grpcmetrics.NewFanoutHandler(internalStatsHandler, pmStatsHandler)

    var opts []grpc.ServerOption

    opts = append(opts, grpc.StatsHandler(statsHandler))


    grpcServer := grpc.NewServer(opts...)
```
