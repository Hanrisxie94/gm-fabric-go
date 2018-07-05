package discovery

import (
	"context"
	"io"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	dv2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Control is a config object so Fabric services can communicate with the Grey Matter envoy management server
type Control struct {
	URL           string   // URL to Envoy management server ex: control.deciphernow.com:10219
	Region        string   // Envoy region/node that will initiate communication with a Fabric service
	ResourceNames []string // List of Envoy resource names to subscribe to
	ResourceType  string
}

// NewSession will dial a new grpc session to the ADS management server
func (a *Control) NewSession(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Dial the grpc envoy management server
	conn, err := grpc.Dial(a.URL, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Fetch consumes a channel and will publish envoy resources to it so services can consume
// You'll want to start this in a go routine and just read the channels passed in as params. It will block
func (a *Control) Fetch(resources chan *types.Any, errs chan error) {
	conn, err := a.NewSession(grpc.WithInsecure())
	if err != nil {
		errs <- errors.Wrap(err, "error when fetching session")
		close(resources)
		return
	}

	// Create our cluster discovery client from the grpc conn
	client := dv2.NewAggregatedDiscoveryServiceClient(conn)

	// Fetch our stream
	stream, err := client.StreamAggregatedResources(context.Background())
	if err != nil {
		errs <- errors.Wrap(err, "error when retreiving stream from client")
		close(resources)
		return
	}

	ctx := stream.Context()
	log.Println(ctx)

	done := make(chan bool, 1)

	// Make requests for clusters
	if err := stream.Send(&v2.DiscoveryRequest{
		Node: &core.Node{
			Id: a.Region,
		},
		ResourceNames: a.ResourceNames,
		TypeUrl:       a.ResourceType,
	}); err != nil {
		errs <- errors.Wrap(err, "error when sending initial stream request")
	}

	// Kick off a go routine to watch the stream
	go func() {
		for {
			resp, err := stream.Recv()
			// If we get an EOF we can assume nothing is left to stream (Envoy ADS should never send an EOF)
			if err == io.EOF {
				errs <- errors.New("received EOF from stream")
				close(done)
			}
			if resp != nil {
				for _, resource := range resp.Resources {
					resources <- &resource
				}

				// Acknowledge the successful response from Envoy ADS
				go ack(stream, resp, a.Region, errs)
			}
		}
	}()

	// Watch the stream context
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			errs <- errors.Wrap(err, "stream context error received")
		}
		close(done)
	}()

	<-done
}

// Acknowledge the reception of a new version of cache resources from Envoy ADS
func ack(stream v2.ClusterDiscoveryService_StreamClustersClient, resp *v2.DiscoveryResponse, region string, errs chan error) {
	// Construct our ACK request and send that through the stream
	if err := stream.Send(&v2.DiscoveryRequest{
		Node: &core.Node{
			Id: region,
		},
		VersionInfo:   resp.GetVersionInfo(),
		ResponseNonce: resp.GetNonce(),
	}); err != nil {
		errs <- errors.Wrap(err, "error when sending ack response")
	}
}
