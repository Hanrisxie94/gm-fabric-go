package discovery

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

// Control is a config object so Fabric services can communicate with the Grey Matter envoy management server
type Control struct {
	URL           string   // URL to Envoy management server ex: control.deciphernow.com:10219
	Region        string   // Envoy region/node that will initiate communication with a Fabric service
	ResourceNames []string // List of Envoy resource names to subscribe to
	ResourceType  string
}

// Fetch consumes a channel and will publish envoy resources to it so services can consume
// You'll want to start this in a go routine and just read the channels passed in as params. It will block
func (a Control) Fetch(resources chan *types.Any, errors chan error) {
	// Dial the grpc envoy management server
	conn, err := grpc.Dial(a.URL, grpc.WithInsecure())
	if err != nil {
		errors <- err
		close(resources)
	}

	// Create our cluster discovery client from the grpc conn
	client := v2.NewClusterDiscoveryServiceClient(conn)
	// Fetch our stream

	stream, err := client.StreamClusters(context.Background())
	if err != nil {
		errors <- err
		close(resources)
	}

	ctx := stream.Context()
	done := make(chan bool, 1)

	go func() {
		for {
			// Make requests for clusters
			req := v2.DiscoveryRequest{
				Node: &core.Node{
					Id: a.Region,
				},
				ResourceNames: a.ResourceNames,
				TypeUrl:       a.ResourceType,
			}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("can not send: %v", err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	// Kick off a go routine to watch the stream
	go func() {
		for {
			resp, err := stream.Recv()
			// If we get an EOF we can assume nothing is left to stream (Envoy ADS should never send an EOF)
			if err == io.EOF {
				close(done)
			}

			for _, resource := range resp.Resources {
				resources <- &resource
			}
		}
	}()

	// Watch the stream context
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Printf("Stream deadline passed: %v\n", err)
		}
		close(done)
	}()

	<-done
	log.Printf("finished fetching clusters")
}
