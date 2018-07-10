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

package discovery

import (
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
)

func TestFetch(t *testing.T) {
	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	timeout := time.After(1 * time.Second)
	done := make(chan bool, 1)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		t.Error(err)
	}

	// Start our ADS resource stream
	go sess.Fetch(clusters, errs)

	// Watch our ADS resource stream
	go func() {
		var c v2.Cluster
		for {
			select {
			case cluster := <-clusters:
				if err := types.UnmarshalAny(cluster, &c); err != nil {
					t.Error(err)
					close(done)
				}
			case err := <-errs:
				t.Error(err)
				close(done)
			case <-timeout:
				close(done)
			}
		}
	}()

	// Block until we are finished watching
	<-done
}

// go test -bench=. -benchmem
func BenchmarkFetch(b *testing.B) {
	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	timeout := time.After(1 * time.Second)
	done := make(chan bool, 1)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		b.Error(err)
	}

	// Start our ADS resource stream
	go sess.Fetch(clusters, errs)

	// Watch our ADS resource stream
	go func() {
		var c v2.Cluster
		for {
			select {
			case cluster := <-clusters:
				if err := types.UnmarshalAny(cluster, &c); err != nil {
					b.Error(err)
					close(done)
				}
			case err := <-errs:
				b.Error(err)
				close(done)
			case <-timeout:
				close(done)
			}
		}
	}()

	// Block until we are finished watching
	<-done
}
