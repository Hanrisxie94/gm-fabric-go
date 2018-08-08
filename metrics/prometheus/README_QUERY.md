# Fabric Prometheus Metrics: Queries

## Instructions

https://prometheus.io/docs/prometheus/latest/querying/basics/
https://prometheus.io/docs/practices/histograms/

## Metrics Collected

* http_request_duration_seconds (vector)
* http_request_size_bytes (vector)
* http_response_size_bytes (vector)
* tls_requests (total)
* non_tls_requests (total)

## Vector partitioning

Each of the vector metrics can be queried by partition

* key (default is URI)
* method (GET, POST, etc)
* status (200, etc)

## Provided with all metrics

Prometheus provides these queryible fields with all metrics

* instance (host id)
* job ('-job-name' from prometheus.yml)



So you can run a query like this:

```
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

