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

// RegistrarOptions stands as a basic configuration object for service registration
type RegistrarOptions struct {
	Check   CheckConfig   `json:"check_config" toml:"check_config" yaml:"check_config"`
	Service ServiceConfig `json:"service_config" toml:"service_config" yaml:"service_config"`
	Tags    []string      `json:"tags" toml:"tags" yaml:"tags"`
}

// RegistrarOption follows functional opts pattern
type RegistrarOption func(*RegistrarOptions)

// Registrar is a high level interface that provides easy service discovery, registration,and deregistration
type Registrar interface {
	// Main method for service registration to consul
	Register(name string, tags []string, port int) error
	// Deregister is a helper method to remvoe a service from consul after it dies
	Deregister(name string) error
}

// ServiceConfig is the basic service configuration object
type ServiceConfig struct {
	Name string `json:"service_name" toml:"name" yaml:"name"`
	Host string `json:"service_host" toml:"host" yaml:"host"`
	Port int    `json:"service_port" toml:"port" yaml:"port"`
}

// CheckConfig is the object you pass to Register that provides configuration option to the Consul Service Check
type CheckConfig struct {
	Protocol string `json:"protocol" toml:"protocol" yaml:"protocol"`     // http - https - tcp
	Interval string `json:"interval" toml:"interval" yaml:"interval"`     // how often you want the health check to run
	Timeout  string `json:"timeout" toml:"timeout" yaml:"timeout"`        // how long the health check should wait before it times out
	Endpoint string `json:"api_endpoint" toml:"endpoint" yaml:"endpoint"` // service endpoint the health check will ping
}

// WithServiceConfig follows functional opts pattern to inject a service config object into the service registration
func WithServiceConfig(s ServiceConfig) RegistrarOption {
	return func(o *RegistrarOptions) {
		o.Service = s
	}
}

// WithCheckConfig follows functional opts pattern to inject a Check configuration object into service registration
func WithCheckConfig(c CheckConfig) RegistrarOption {
	return func(o *RegistrarOptions) {
		o.Check = c
	}
}

// WithTags follows functional opts pattern to inject Tags into service registration
func WithTags(tags []string) RegistrarOption {
	return func(o *RegistrarOptions) {
		o.Tags = tags
	}
}
