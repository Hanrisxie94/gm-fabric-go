package impersonation

import (
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/rs/zerolog"
)

// Whitelist object represents a list of servers to whitelist
type Whitelist struct {
	Servers []string `json:"server_whitelist"`
}

// NewWhitelist will construct a whitelist object and return
func NewWhitelist(servers []string) Whitelist {
	return Whitelist{
		Servers: servers,
	}
}

// CanImpersonate will check the server whitelist to see if the EXTERNAL_SYS_DN lives in it's store.
// If not it will return false and the impersonation request should be denied
func CanImpersonate(caller Caller, whitelist Whitelist) bool {
	var isValid bool
	for _, server := range whitelist.Servers {
		isValid = validate(caller.ExternalSystemDistinguishedName, server)
	}

	return isValid
}

func validate(sysDN, ws string) bool {
	if strings.Compare(strings.Trim(sysDN, " "), strings.Trim(ws, " ")) == 0 {
		return true
	}
	return false
}

// ValidateCaller will check to see if the server is on the whitelist and if not, block the request
func ValidateCaller(whitelist Whitelist, logger zerolog.Logger) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cert *x509.Certificate
			if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
				cert = r.TLS.PeerCertificates[0]
			}
			caller := GetCaller(r.Header.Get(USER_DN), r.Header.Get(EXTERNAL_SYS_DN), cert)

			if CanImpersonate(caller, whitelist) {
				next.ServeHTTP(w, r)
			} else {
				logger.Error().Str(EXTERNAL_SYS_DN, caller.ExternalSystemDistinguishedName).Msg("Server not on authorized whitelist")
				return
			}
		})
	})
}
