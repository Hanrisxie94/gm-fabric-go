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
	"context"
	"io"
	"net"
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func TestFetch(t *testing.T) {
	err := createMockADSServer(context.Background())
	if err != nil {
		t.Error(err)
	}

	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	done := make(chan bool, 1)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("localhost:18011"))
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
					break
				}
				break
			case err := <-errs:
				if err == io.EOF {
					close(done)
					break
				}
				t.Error(err)
				break
			}
		}
	}()

	// Block until we are finished watching
	<-done
}

// go test -bench=. -benchmem
func BenchmarkFetch(b *testing.B) {
	err := createMockADSServer(context.Background())
	if err != nil {
		b.Error(err)
	}

	// Create a buffered channel
	clusters := make(chan *types.Any, 1)
	errs := make(chan error, 1)
	done := make(chan bool, 1)

	// Create a control object with necessary metadata
	sess, err := NewDiscoverySession(WithRegion("region-1"), WithResourceType(cache.ClusterType), WithLocation("localhost:18011"))
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
				}
			case err := <-errs:
				if err == io.EOF {
					close(done)
					break
				}
				b.Error(err)
			}
		}
	}()

	// Block until we are finished watching
	<-done
}

func createMockADSServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":18011")
	if err != nil {
		return err
	}

	go func() {
		s := Server{}
		grpcServer := grpc.NewServer()
		discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, &s)

		if e := grpcServer.Serve(lis); e != nil {
			err = e
		}
	}()

	return err
}

type Server struct{}

func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	// Basic cluster info
	clusters := []*v2.Cluster{
		&v2.Cluster{Name: "cluster_1"},
		&v2.Cluster{Name: "cluster_2"},
		&v2.Cluster{Name: "cluster_3"},
	}

	var resources []types.Any
	for _, cluster := range clusters {
		c, err := types.MarshalAny(cluster)
		if err != nil {
			return err
		}
		resources = append(resources, *c)
	}

	// Send one streamg
	stream.Send(&v2.DiscoveryResponse{
		VersionInfo: "1",
		TypeUrl:     cache.ClusterType,
		Resources:   resources,
	})

	return nil
}

func (s *Server) IncrementalAggregatedResources(_ discovery.AggregatedDiscoveryService_IncrementalAggregatedResourcesServer) error {
	return errors.New("not implemented")
}
