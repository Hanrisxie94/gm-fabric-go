# Discovery
Package discovery enables easy service discovery while following the Envoy Management Server specification

## Usage

### Endpoint

```go
package main

import (
    "github.com/deciphernow/gm-fabric-go/discovery"
    "github.com/deciphernow/gm-fabric-go/tlsutil"
    "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func main() {
    req := v2.DiscoveryRequest{}
    e := discovery.NewEndpoint(&req)

    locations := make(map[string][]string)
    locations["endpoint"] = []string{"endpoint_service:8080"}
    err := discovery.AnnounceGRPC(e, discovery.WithLocation(locations))
    if err != nil {
        log.Fatal(err)
    }    
}
```

Optionally you can provide a `*tls.Config` option if you wish to speak 2-Way SSL with the Endpoint Discovery System

```go
cnf, err := tlsutil.NewTLSClientConfig("./cacert.crt", "./cert.crt", "./key.key", "test-server-cn")
if err != nil {
    log.Fatal(err)
}

err := discovery.AnnounceGRPC(e, discovery.WithLocation(locations), discovery.WithTLSConfig(cnf))
if err != nil {
    log.Fatal(err)
}
```

We also support announement over `HTTP/2` and `REST` if you wish to not go directly through gRPC.

### Cluster

```go
package main

import (
    "github.com/deciphernow/gm-fabric-go/discovery"
    "github.com/deciphernow/gm-fabric-go/tlsutil"
    "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func main() {
    req := v2.DiscoveryRequest{}
    c := discovery.NewCluster(&req)

    locations := make(map[string][]string)
    locations["endpoint"] = []string{"endpoint_service:8080"}
    err := discovery.AnnounceGRPC(c, discovery.WithLocation(locations))
    if err != nil {
        log.Fatal(err)
    }    
}
```

Optionally you can provide a `*tls.Config` option if you wish to speak 2-Way SSL with the Endpoint Discovery System

```go
cnf, err := tlsutil.NewTLSClientConfig("./cacert.crt", "./cert.crt", "./key.key", "test-server-cn")
if err != nil {
    log.Fatal(err)
}

err := discovery.AnnounceGRPC(e, discovery.WithLocation(locations), discovery.WithTLSConfig(cnf))
if err != nil {
    log.Fatal(err)
}
```

We also support announement over `HTTP/2` and `REST` if you wish to not go directly through gRPC.

### Route

```go
package main

import (
    "github.com/deciphernow/gm-fabric-go/discovery"
    "github.com/deciphernow/gm-fabric-go/tlsutil"
    "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func main() {
    req := v2.DiscoveryRequest{}
    r := discovery.NewRoute(&req)

    locations := make(map[string][]string)
    locations["endpoint"] = []string{"endpoint_service:8080"}
    err := discovery.AnnounceGRPC(r, discovery.WithLocation(locations))
    if err != nil {
        log.Fatal(err)
    }    
}
```

Optionally you can provide a `*tls.Config` option if you wish to speak 2-Way SSL with the Endpoint Discovery System

```go
cnf, err := tlsutil.NewTLSClientConfig("./cacert.crt", "./cert.crt", "./key.key", "test-server-cn")
if err != nil {
    log.Fatal(err)
}

err := discovery.AnnounceGRPC(e, discovery.WithLocation(locations), discovery.WithTLSConfig(cnf))
if err != nil {
    log.Fatal(err)
}
```

We also support announement over `HTTP/2` and `REST` if you wish to not go directly through gRPC.

### Listener

```go
package main

import (
    "github.com/deciphernow/gm-fabric-go/discovery"
    "github.com/deciphernow/gm-fabric-go/tlsutil"
    "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func main() {
    req := v2.DiscoveryRequest{}
    l := discovery.NewListener(&req)

    locations := make(map[string][]string)
    locations["endpoint"] = []string{"endpoint_service:8080"}
    err := discovery.AnnounceGRPC(l, discovery.WithLocation(locations))
    if err != nil {
        log.Fatal(err)
    }    
}
```

Optionally you can provide a `*tls.Config` option if you wish to speak 2-Way SSL with the Endpoint Discovery System

```go
cnf, err := tlsutil.NewTLSClientConfig("./cacert.crt", "./cert.crt", "./key.key", "test-server-cn")
if err != nil {
    log.Fatal(err)
}

err := discovery.AnnounceGRPC(e, discovery.WithLocation(locations), discovery.WithTLSConfig(cnf))
if err != nil {
    log.Fatal(err)
}
```
