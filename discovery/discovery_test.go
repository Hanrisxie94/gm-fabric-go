package discovery

import (
	"fmt"
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
	done := make(chan bool, 1)
	timeout := time.After(1 * time.Second)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		t.Error(err)
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
					fmt.Println(c.String())
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
	done := make(chan bool, 1)
	timeout := time.After(1 * time.Second)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("control.deciphernow.com:10219"))
	if err != nil {
		b.Error(err)
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
