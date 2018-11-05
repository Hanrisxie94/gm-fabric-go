# Metrics Instrumentation

## Observations

We capture thoughput over a short timespan determined by ```throughput_timeout_duration```

* BeginTime is the earliest point that we can store a timestamp
* RequestTime is the time when an HTTP request is available, with all headers received
* ResponseTime is the time when an HTTP response is ready to send
* EndTime is the time when the transaction is completely ended
* InCaptureTime is the time when InWireLength is captured
* InWireLength is the amount of data received from the request body at InCaptureTime
* OutCaptureTime is the time when OutWireLength is captured
* OutWireLength is the amount of data sent from the response body at OutCaptureTime

## Computations

* Latency ms = RequestTime - BeginTime 
* InThroughput bytes/sec = InWireLength / (InCaptureTime - RequestTime) 
* OutThroughput bytes/sec = OutWireLength / (OutCaptureTime - ResponseTime)

## Metrics Server Output

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
