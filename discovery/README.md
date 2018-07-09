# Discovery
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/discovery)

A package for discovery Envoy resources from the Aggregate Discovery Service or Envoy Management Server

## Usage
Here is a basic example of fetching Clusters (an Envoy Resource Type) from the Aggregate Discovery Service using common Go concurrency patterns and this package:
```go
package main

import (
	"testing"
	"time"
    "log"

	"github.com/deciphernow/gm-fabric-go/discovery"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
)

func main() {
	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	done := make(chan bool, 1)

	// Setting a timeout is optional. The stream will stay open infinitely if none is set
	timeout := time.After(10 * time.Second)

	// Create a control object with necessary metadata
	sess, err := discovery.NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		log.Fatal(err)
	}

	// Start our ADS resource stream
	go sess.Fetch(clusters, errs)


	// Watch our ADS resource stream
	go func() {
	    for {
	        select {
	        case cluster := <-clusters:
	            var c v2.Cluster
	            if err := types.UnmarshalAny(cluster, &c); err != nil {
	                log.Println(err)
	                close(done)
	            }
	        case err := <-errs:
	            log.Println(err)
	            close(done)
	        case <-timeout:
	            close(done)
	        }
	    }
	}()

	// Block until we are finished watching
	<-done    
}
```

## Performance
Current performance of the Discovery package:
```
goos: linux
goarch: amd64
pkg: github.com/deciphernow/gm-fabric-go/discovery
BenchmarkFetch-32              1        1000193976 ns/op           29296 B/op        289 allocs/op
PASS
ok      github.com/deciphernow/gm-fabric-go/discovery   2.011s
```
