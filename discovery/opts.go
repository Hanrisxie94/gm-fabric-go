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
