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
	done := make(chan bool, 1)
	timeout := time.After(1 * time.Second)

	// Create a control object with necessary metadata
	c := &Control{
		URL:          "control.deciphernow.com:10219", // URL to an ADS instance
		Region:       "region-1",                      // Region ADS is apart of
		ResourceType: cache.ClusterType,               // Envoy resource type we want
	}

	// Start our ADS resource stream
	go c.Fetch(clusters, errs)

	// Watch our ADS resource stream
	go func() {
		for {
			select {
			case cluster := <-clusters:
				var c v2.Cluster
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
	done := make(chan bool, 1)
	timeout := time.After(1 * time.Second)

	// Create a control object with necessary metadata
	c := &Control{
		URL:          "control.deciphernow.com:10219", // URL to an ADS instance
		Region:       "region-1",                      // Region ADS is apart of
		ResourceType: cache.ClusterType,               // Envoy resource type we want
	}

	// Start our ADS resource stream
	go c.Fetch(clusters, errs)

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
