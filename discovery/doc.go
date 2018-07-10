// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package discovery provides an easy way to communicate with an Envoy Management Server

A contrived example:
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
*/
package discovery
