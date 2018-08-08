# Impersonation
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/impersonation)
Package for 2-Way SSL impersonation

## Usage
A `Caller` is a description of the (possibly impersonated) identity of the initiator of an incoming request. An important assumption is that there exists a trustworthy, TLS-terminating proxy between the replying service and the application making the request.

The proxy is expected to provide two headers:
- USER_DN
- EXTERNAL_SYS_DN

Typical usage will look something like:
```go
var req *http.Request

var cert *x509.Certificate
if req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
	cert = req.TLS.PeerCertificates[0]
}

caller := GetCaller(
    req.Header.Get(impersonation.USER_DN),
    req.Header.Get(impersonation.S_CLIENT_S_DN),
    req.Header.Get(impersonation.EXTERNAL_SYS_DN),
    cert,
)
```
An x509 certificate can be provided to use as a fallback when a USER_DN header is not present, in which case the DN
from the cert will be used. This should only be necessary in the unlikely scenario where you need to allow an
application to bypass the trusted proxy and establish a direct TLS connection to your service.

## Server Impersonation
Users who wish to enable the AccessControlList (ACL) middleware can inject the `ValidateCaller` function into their middleware stack like so:
```go
package main

import (
	"github.com/deciphernow/gm-fabric-go/impersonation"
	"github.com/deciphernow/gm-fabric-go/middleware"
)

func main() {
	...

	m := []middleware.Middleware{
		middleware.MiddlewareFunc(cors.AllowAll().Handler),
		middleware.MiddlewareFunc(hlog.NewHandler(logger)),
		middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("path", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("Access")
		})),
		middleware.MiddlewareFunc(hlog.UserAgentHandler("user_agent")),
	}

	// Create your server whitelist
	whitelist := impersonation.NewWhitelist([]string{
		"uid=server1,ou=Server,dc=example,dc=com",
		"uid=server2,ou=Server,dc=example,dc=com",
		"uid=server3,ou=Server,dc=example,dc=com",
	})
	if enable_acl {
		m = append(m, impersonation.ValidateCaller(whitelist))
	}

	...
}
```
This will wrap every `http.HandlerFunc` registered in your `http.Server` and provide a high-level control module within your middleware stack. Alternatively you may wish to not inject middleware. If that is the case you can use the `impersonation.CanImpersonate` function within an `http.HandlerFunc` to validate the Caller object against your server whitelist.

## Info
**USER_DN** - The effective (possibly impersonated) Distinguished Name of requesting application

**S_CLIENT_S_DN** - The Distinguished Name taken from the system certificate

**EXTERNAL_SYS_DN** - The Distinguished Name taken from the external system certificate (originally inside s_client_s_dn)
