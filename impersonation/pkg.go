package impersonation

import (
	"crypto/x509"

	"github.com/deciphernow/gm-fabric-go/tlsutil"
)

// Caller provides the distinguished names obtained from specific request
// headers and peer certificate if called directly
type Caller struct {
	// DistinguishedName is the unique identity of a user
	DistinguishedName string
	// UserDistinguishedName holds the value passed in header USER_DN
	UserDistinguishedName string
	// SystemDistinguishedName holds the value passed in header S_CLIENT_S_DN
	SystemDistinguishedName string
	// ExternalSystemDistinguishedName holds the value passed in header EXTERNAL_SYS_DN
	ExternalSystemDistinguishedName string
	// CommonName is the CN value part of the DistinguishedName
	CommonName string
}

var (
	USER_DN         = "USER_DN"
	EXTERNAL_SYS_DN = "EXTERNAL_SYS_DN"
	SSL_CLIENT_S_DN = "SSL_CLIENT_S_DN"
)

/*
GetCaller creates a description of the (possibly impersonated) identity of the initiator of an incoming request.

An important assumption is that there exists a trustworthy, TLS-terminating proxy between the replying service
and the application making the request.

The proxy is expected to provide two headers:

USER_DN
	The effective (possibly impersonated) Distinguished Name of requesting application
S_CLIENT_S_DN
    The Distinguished Name taken from the system certificate
EXTERNAL_SYS_DN
	The Distinguished Name taken from the external system certificate (originally inside s_client_s_dn)

An x509 certificate can be provided to use as a fallback when a USER_DN header is not present, in which case the DN
from the cert will be used. This should only be necessary in the unlikely scenario where you need to allow an
application to bypass the trusted proxy and establish a direct TLS connection to your service.

Typical usage will look something like so:

	// This would usually be declared as a parameter in the definition of e.g. a http.Handler.
	var req *http.Request

	// Get the cert from the request.
	//
	// Note that req.TLS will be nil if you're not using the stdlib's
	// impelemtnation of TLS (e.g. if you're using spacemonkeygo/openssl).
	// See https://github.com/golang/go/issues/14891
	var cert *x509.Certificate
	if req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
		cert = req.TLS.PeerCertificates[0]
	}

	caller := GetCaller(
		req.Header.Get(impersonation.USER_DN),
		req.Header.Get(impersonation.SSL_CLIENT_S_DN),
		req.Header.Get(impersonation.EXTERNAL_SYS_DN),
		cert,
	)
*/
func GetCaller(userDN, sysDn, externalSysDN string, cert *x509.Certificate) Caller {
	var caller Caller
	caller.UserDistinguishedName = userDN
	caller.ExternalSystemDistinguishedName = externalSysDN
	caller.SystemDistinguishedName = sysDn
	if caller.UserDistinguishedName != "" {
		caller.DistinguishedName = caller.UserDistinguishedName
	} else {
		if cert != nil {
			caller.DistinguishedName = tlsutil.GetDistinguishedName(cert)
		}
	}
	caller.DistinguishedName = tlsutil.GetNormalizedDistinguishedName(caller.DistinguishedName)
	caller.CommonName = tlsutil.GetCommonName(caller.DistinguishedName)
	return caller
}
