# Fabric Prometheus Metrics: Queries

## Instructions

https://prometheus.io/docs/prometheus/latest/querying/basics/

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

Prometheus provides these queryible fields wiht all metrics

* instance (host id)
* job ('-job-name' from prometheus.yml)



So you can run a query like this:

```
avg(rate(http_request_duration_seconds_count[5m]))  by (job)
```

The request duration summary reports quantiles

```
http_request_duration_seconds{quantile="0.9999"}
```

results

```text
Element	Value
http_request_duration_seconds{instance="envoymesh:60001",job="store",key="all",quantile="0.9999",status="0"}	2.0063339
http_request_duration_seconds{instance="envoymesh:60001",job="store",key="function/CatalogStream",method="gRPC",quantile="0.9999",status="0"}	2.0016597
http_request_duration_seconds{instance="envoymesh:60001",job="store",key="function/OrderItem",method="gRPC",quantile="0.9999",status="0"}	2.0063339
http_request_duration_seconds{instance="envoymesh:60002",job="store-http",key="/acme/services/catalog",method="GET",quantile="0.9999",status="200"}	1.9931261
http_request_duration_seconds{instance="envoymesh:60002",job="store-http",key="/acme/services/order",method="POST",quantile="0.9999",status="200"}	2.0089295
http_request_duration_seconds{instance="envoymesh:60002",job="store-http",key="all",quantile="0.9999",status="200"}	2.0089295
```