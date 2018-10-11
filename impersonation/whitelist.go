package impersonation

import (
	"crypto/x509"
	"net/http"
	"strings"
	"unicode"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/rs/zerolog"
)

// Whitelist object represents a list of servers to whitelist
type Whitelist struct {
	servers []string
}

// NewWhitelist will construct a whitelist object and return
func NewWhitelist(servers []string) Whitelist {
	srvrs := make([]string, 0)
	for _, s := range servers {
		s = normalize(s)
		srvrs = append(srvrs, s)
	}

	return Whitelist{
		servers: srvrs,
	}
}

// CanImpersonate will check the server whitelist to see if the EXTERNAL_SYS_DN lives in it's store.
// If not it will return false and the impersonation request should be denied
func CanImpersonate(caller Caller, whitelist Whitelist) bool {
	return validate(caller, whitelist)
}

func validate(caller Caller, whitelist Whitelist) bool {
	if caller.SystemDistinguishedName == "" && caller.ExternalSystemDistinguishedName == "" {
		return false
	} else if caller.ExternalSystemDistinguishedName == "" {
		for _, server := range whitelist.servers {
			if normalize(caller.SystemDistinguishedName) == server {
				return true
			}
		}
		return false
	}
	return validateExternalSystem(caller, whitelist)
}

func validateExternalSystem(caller Caller, whitelist Whitelist) bool {
	var foundExternal bool
	var foundInternal bool
	for _, s := range whitelist.servers {
		if s == normalize(caller.ExternalSystemDistinguishedName) {
			foundExternal = true
		}
		if s == normalize(caller.SystemDistinguishedName) {
			foundInternal = true
		}
	}
	return foundInternal && foundExternal
}

// ValidateCaller will check to see if the server is on the whitelist and if not, block the request
func ValidateCaller(whitelist Whitelist, logger zerolog.Logger) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cert *x509.Certificate
			if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
				cert = r.TLS.PeerCertificates[0]
			}
			caller := GetCaller(r.Header.Get(USER_DN), r.Header.Get(SSL_CLIENT_S_DN), r.Header.Get(EXTERNAL_SYS_DN), cert)

			if CanImpersonate(caller, whitelist) {
				logger.Info().Str(EXTERNAL_SYS_DN, caller.ExternalSystemDistinguishedName).Str(SSL_CLIENT_S_DN, caller.SystemDistinguishedName).Str(USER_DN, caller.UserDistinguishedName).Msg("Impersonation successful")
				next.ServeHTTP(w, r)
			} else {
				logger.Error().Str(EXTERNAL_SYS_DN, caller.ExternalSystemDistinguishedName).Str(SSL_CLIENT_S_DN, caller.SystemDistinguishedName).Str(USER_DN, caller.UserDistinguishedName).Msg("Server not on authorized whitelist")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		})
	})
}

// normalize will remove all whitespace from all DNs in the server list
func normalize(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	return strings.ToLower(s)
}
