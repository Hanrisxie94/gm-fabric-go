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

// Options are configuration settings for the discovery object
type Options struct {
	URL           string   // URL to Envoy management server ex: control.deciphernow.com:10219
	Region        string   // Envoy region/node that will initiate communication with a Fabric service
	ResourceNames []string // List of Envoy resource names to subscribe to
	ResourceType  string
}

// Option follows the functional opts pattern
type Option func(*Options)

// WithLocation will inject a URL string into the configuration object
func WithLocation(url string) Option {
	return func(o *Options) {
		o.URL = url
	}
}

// WithRegion will inject a region string to the configuration object
func WithRegion(region string) Option {
	return func(o *Options) {
		o.Region = region
	}
}

// WithResourceNames will inject a list of resources the user wants to watch into the configuration object
func WithResourceNames(names []string) Option {
	return func(o *Options) {
		o.ResourceNames = names
	}
}

// WithResourceType will inject the specific resource type that a user wants to stream
func WithResourceType(resource string) Option {
	return func(o *Options) {
		o.ResourceType = resource
	}
}
