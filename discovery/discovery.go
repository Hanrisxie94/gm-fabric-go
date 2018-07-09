package discovery

import (
	"context"
	"io"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	dv2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Discovery is a config object so Fabric services can communicate with the Grey Matter envoy management server
type Discovery struct {
	Session *grpc.ClientConn
	Options Options
}

// NewDiscoverySession creates a Control ADS grpc session
func NewDiscoverySession(opts ...Option) (*Discovery, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	sess, err := options.NewSession(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Discovery{
		Session: sess,
		Options: options,
	}, nil
}

// NewSession will dial a new grpc session to the ADS management server
func (a *Options) NewSession(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Dial the grpc envoy management server
	conn, err := grpc.Dial(a.URL, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Fetch consumes a channel and will publish envoy resources to it so services can consume
// You'll want to start this in a go routine and just read the channels passed in as params. It will block
func (a *Discovery) Fetch(resources chan *types.Any, errs chan error) {
	// Create our ADS client from the grpc conn
	client := dv2.NewAggregatedDiscoveryServiceClient(a.Session)

	// Fetch our stream
	stream, err := client.StreamAggregatedResources(context.Background())
	if err != nil {
		errs <- errors.Wrap(err, "error when retreiving stream from client")
		return
	}

	ctx := stream.Context()
	done := make(chan bool, 1)

	// Make requests for clusters
	if err := stream.Send(&v2.DiscoveryRequest{
		Node: &core.Node{
			Id: a.Options.Region,
		},
		ResourceNames: a.Options.ResourceNames,
		TypeUrl:       a.Options.ResourceType,
	}); err != nil {
		errs <- errors.Wrap(err, "error when sending initial stream request")
		return
	}

	// Kick off a go routine to watch the stream
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				errs <- errors.New("received EOF from stream")
				close(done)
				return
			} else if err != nil {
				errs <- errors.Wrap(err, "failed in resource reception")
				close(done)
				return
			}
			if resp != nil {
				for _, resource := range resp.Resources {
					resources <- &resource
				}

				// Acknowledge the successful response from Envoy ADS
				go ack(stream, resp, a.Options.Region, errs)
			}
		}
	}()

	// Watch the stream context
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil && err != context.Canceled {
			errs <- errors.Wrap(err, "stream context error received")
			close(done)
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
		TypeUrl:       resp.GetTypeUrl(),
	}); err != nil {
		errs <- errors.Wrap(err, "error when sending acknowledgement response")
	}
}
