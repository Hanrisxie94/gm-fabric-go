package discovery

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Cluster ...
type Cluster struct {
	Request *v2.DiscoveryRequest
}

// NewCluster ...
func NewCluster(req *v2.DiscoveryRequest) Cluster {
	return Cluster{
		Request: req,
	}
}

// Announce ...
func (c *Cluster) Announce(opts ...AnnounceOption) error {
	var options AnnounceOptions
	for _, o := range opts {
		o(&options)
	}

	var locations []string
	if options.Location == nil {
		locations = DefaultLocations["cluster"]
	} else {
		locations = options.Location["cluster"]
	}

	body, err := c.Request.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal discovery request object")
	}

	var req *http.Request

	for _, location := range locations {
		// Parse out the host and port from the location string
		l := strings.Split(location, ":")
		host := l[0]
		port := l[1]

		// Add the Envoy spec URI to the location
		// /v2/discovery:endpoints
		location += "/v2/discovery/endpoint"

		req, err = http.NewRequest("POST", location, bytes.NewBuffer(body))
		if err != nil {
			return errors.Wrap(err, "failed to create request")
		}
		req.Close = true

		if options.TLSConfig != nil {
			client := &http.Client{
				Transport: &http.Transport{
					DialTLS: func(network, address string) (net.Conn, error) {
						return tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), options.TLSConfig)
					},
				},
			}

			res, err := client.Do(req)
			if err != nil {
				return errors.Wrap(err, "failed to send request to EDS")
			}
			defer res.Body.Close()
			err = checkResponseCodes(res.StatusCode)
			if err != nil {
				return err
			}
		} else {
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Wrap(err, "failed to use default http client")
			}
			defer res.Body.Close()
			err = checkResponseCodes(res.StatusCode)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AnnounceGRPC ...
func (c *Cluster) AnnounceGRPC(opts ...AnnounceOption) error {
	var options AnnounceOptions
	for _, o := range opts {
		o(&options)
	}

	var locations []string
	if options.Location == nil {
		locations = DefaultLocations["endpoint"]
	} else {
		locations = options.Location["endpoint"]
	}

	var conn *grpc.ClientConn
	var dialOpts []grpc.DialOption
	var err error
	for _, location := range locations {
		if options.TLSConfig != nil {
			creds := credentials.NewTLS(options.TLSConfig)
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		} else {
			dialOpts = append(dialOpts, grpc.WithInsecure())
		}

		conn, err = grpc.Dial(location, dialOpts...)
		if err != nil {
			return errors.Wrap(err, "failed to dial grpc server")
		}

		client := v2.NewEndpointDiscoveryServiceClient(conn)

		_, err := client.FetchEndpoints(context.Background(), c.Request, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
