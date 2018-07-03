package discovery

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
)

func TestFetch(t *testing.T) {
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	done := make(chan bool, 1)
	timeout := time.After(3 * time.Second)

	c := Control{
		URL:          "control.deciphernow.com:10219",
		Region:       "region-1",
		ResourceType: cache.ClusterType,
	}

	go c.Fetch(clusters, errs)

	go func() {
		for {
			select {
			case cluster := <-clusters:
				var c v2.Cluster
				if err := types.UnmarshalAny(cluster, &c); err != nil {
					t.Fatal(err)
					close(done)
				}
				fmt.Println(c.String())
			case err := <-errs:
				t.Fatal(err)
				close(done)
			case <-timeout:
				log.Println("stream timeout")
				close(done)
			}
		}
	}()

	<-done
	log.Println("exiting")
}
