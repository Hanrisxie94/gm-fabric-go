package discovery

import "crypto/tls"

var (
	// DefaultLocations options for each discovery management service
	DefaultLocations = map[string][]string{
		"cluster":  nil,
		"service":  nil,
		"endpoint": nil,
		"route":    nil,
		"listener": nil,
	}
)

// Announcer is the base interface to implement for each discovery package
// Cluster, Service, Route, etc... all share a common announcement functionallity just with different parameters
type Announcer interface {
	Announce(opts ...AnnounceOption) error
	AnnounceGRPC(opts ...AnnounceOption) error
}

// AnnounceOptions are options that might help with announcement such as the location of the service, or tls configurations for 2-Way SSL
type AnnounceOptions struct {
	Location  map[string][]string
	TLSConfig *tls.Config
}

// AnnounceOption follows the funtional opts pattern
type AnnounceOption func(*AnnounceOptions)

// WithTLSConfig will pass a tls configuration object into the announcement options enabling 2-Way SSL
func WithTLSConfig(cnf *tls.Config) AnnounceOption {
	return func(o *AnnounceOptions) {
		o.TLSConfig = cnf
	}
}

// WithLocations specificies the location of all discovery management services
func WithLocations(l map[string][]string) AnnounceOption {
	return func(o *AnnounceOptions) {
		o.Location = l
	}
}

// Announce ...
func Announce(a Announcer, opts ...AnnounceOption) error {
	return a.Announce(opts...)
}

// AnnounceGRPC ...
func AnnounceGRPC(a Announcer, opts ...AnnounceOption) error {
	return a.AnnounceGRPC(opts...)
}
