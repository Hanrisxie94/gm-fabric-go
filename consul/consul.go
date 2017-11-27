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

package consul

import (
	"strconv"

	"github.com/deciphernow/gm-fabric-go/dbutil"
	"github.com/hashicorp/consul/api"
)

// Client is the basic object that a developer can create to enable service discovery/registration/health checks/deregistration
type Client struct {
	*api.Client
}

// Register will add a service into the consul agent and register a check to monitor the service from consul
func (c *Client) Register(opts ...RegistrarOption) (string, error) {
	id := dbutil.CreateHash()
	var r *api.AgentServiceRegistration
	var options RegistrarOptions
	for _, o := range opts {
		o(&options)
	}

	if options.Check == (CheckConfig{}) {
		// Register a service without a check
		r = &api.AgentServiceRegistration{
			// Create a unique hash for the service
			ID: id,
			// Name of service
			Name: options.Service.Name,
			// Tags for consul
			Tags: options.Tags,
			// Port of service
			Port: options.Service.Port,
		}
	} else {
		// Register a service with an http check
		r = &api.AgentServiceRegistration{
			// Create a unique hash for the service
			ID: id,
			// Name of service
			Name: options.Service.Name,
			// Tags for consul
			Tags: options.Tags,
			// Port of service
			Port: options.Service.Port,
			// Register an HTTP check
			Check: &api.AgentServiceCheck{
				Interval: options.Check.Interval,
				Timeout:  options.Check.Timeout,
				HTTP:     options.Check.Protocol + "://" + options.Service.Host + ":" + strconv.Itoa(options.Service.Port) + options.Check.Endpoint,
				Status:   "passing",
			},
		}
	}

	// Register service with consul using Services API (return the registered hash as well)
	return id, c.Agent().ServiceRegister(r)
}

// Deregister will remove a service from the consul agent
func (c *Client) Deregister(id string) error {
	// Deregister the service using the unique hash
	return c.Agent().ServiceDeregister(id)
}

// NewConsulClient will connect to a consul agent with the provided address, and return a client
func NewConsulClient(addr string) (Client, error) {
	// initlialize a default consul config
	config := api.DefaultConfig()
	config.Address = addr

	// create the client
	// NOTE: currently we don't support TLS connections to consul
	client, err := api.NewClient(config)
	if err != nil {
		return Client{}, err
	}

	// send back a client object with a pointer to the consul client
	return Client{
		client,
	}, nil
}
