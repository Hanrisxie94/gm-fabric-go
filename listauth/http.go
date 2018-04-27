// Copyright 2017 Decipher Technology Studios LLC
// go
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
// go
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package listauth

import (
	"fmt"
	"net/http"

	"github.com/deciphernow/gm-fabric-go/listauth/auth"
	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/deciphernow/gm-fabric-go/tlsutil"
	"github.com/rs/zerolog"
)

// HTTPAuthenticate performs a server-side blacklist/whitelist check to
// authenticate a users DN
func HTTPAuthenticate(
	logger zerolog.Logger,
	authorizor auth.Authorizor,
) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return handlerFactory(logger, authorizor, next)
	}
}

func handlerFactory(
	logger zerolog.Logger,
	authorizor auth.Authorizor,
	next http.Handler,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.TLS == nil || r.TLS.PeerCertificates == nil {
			logger.Error().Msg("Unable to get peer certificate")
			http.Error(
				w,
				fmt.Sprintf("Unable to get peer certificate"),
				http.StatusUnauthorized,
			)
			return
		}

		dn := tlsutil.GetDistinguishedName(r.TLS.PeerCertificates[0])

		authorized, err := authorizor.IsAuthorized(dn)
		if err != nil {
			logger.Error().Err(err).Msgf("authorizor.IsAuthorized(%s)", dn)
			http.Error(
				w,
				fmt.Sprintf("error authorizing DN: %s", err),
				http.StatusInternalServerError,
			)
			return
		}
		if !authorized {
			logger.Debug().Str("Unauthorized DN", dn).Msg("HTTPAuthenticate")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
