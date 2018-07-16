# Impersonation
Package for 2-Way SSL impersonation

## Usage
GetCaller creates a description of the (possibly impersonated) identity of the initiator of an incoming request.

An important assumption is that there exists a trustworthy, TLS-terminating proxy between the replying service
and the application making the request.

The proxy is expected to provide two headers:

USER_DN
	The effective (possibly impersonated) Distinguished Name of requesting application
EXTERNAL_SYS_DN
	The Distinguished Name taken from the client certificate

An x509 certificate can be provided to use as a fallback when a USER_DN header is not present, in which case the DN
from the cert will be used. This should only be necessary in the unlikely scenario where you need to allow an
application to bypass the trusted proxy and establish a direct TLS connection to your service.

Typical usage will look something like so:
```go
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
	req.Header.Get(impersonation.EXTERNAL_SYS_DN),
	cert,
)
```
