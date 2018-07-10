# Discovery
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/discovery)

A package for discovery Envoy resources from the Aggregate Discovery Service or Envoy Management Server

## Usage
Here is a basic example of fetching Clusters (an Envoy Resource Type) from the Aggregate Discovery Service using common Go concurrency patterns and this package:
```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/deciphernow/gm-fabric-go/discovery"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	done := make(chan bool, 1)

	// Setting a timeout is optional. The stream will stay open infinitely if none is set
	timeout := time.After(10 * time.Second)

	// Create a control object with necessary metadata
	sess, err := discovery.NewDiscoverySession(discovery.WithRegion("region-1"), discovery.WithResourceType(cache.ListenerType), discovery.WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed when creating discovery session")
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
					logger.Error().Err(err).Msg("failed to unmarshal cluster object")
					close(done)
				}
				fmt.Println(c.String())
			case err := <-errs:
				logger.Error().Err(err).Msg("Received error from error channel")
				close(done)
			case <-timeout:
				logger.Warn().Msg("timed out")
				close(done)
			}
		}
	}()

	// Block until we are finished watching
	<-done
	logger.Info().Msg("exiting")
}
```

The Envoy resource types are located [here](https://github.com/envoyproxy/go-control-plane/blob/master/pkg/cache/resource.go#L32). Because we are subscribing to the Aggregate Discovery model, you are required to provide a resource type URL, even if it is `cache.AnyType`. We recommend asking for your specific resource type otherwise the payload may get very large.

## Performance
Current performance of the Discovery package:
```
goos: darwin
goarch: amd64
pkg: github.com/deciphernow/gm-fabric-go/discovery
BenchmarkFetch-8               1        1000702613 ns/op           14648 B/op        167 allocs/op
PASS
ok      github.com/deciphernow/gm-fabric-go/discovery   2.021s
```
