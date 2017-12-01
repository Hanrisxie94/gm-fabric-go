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

package oauth

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	auth0 "github.com/auth0/go-jwt-middleware"
	"github.com/deciphernow/gm-fabric-go/middleware"
)

/*
NOTE: Middleware loads top down. Always load your signing algorithm first

// Inject the JWT middleware
stack := middleware.Chain(
    // Adding this to the stack will require all API queries to provide a token in the auth http header
    // Always pass the signing algorithm to expect and any necessary supporting data such as a signing secret or key path
    oauth.HTTPAuthenticate(oauth.WithSigningAlg(HS256), oauth.WithHMACSecret("KbtfnXOHH3ezniXIsHYSd4WhZPBXH1vB")),
)

Note: If the token is base64 encoded, it must be decoded before being passed into the middleware
*/

// HTTPAuthenticate performs a server-side OAuth 2.0 flow to authenicate a users JWT
func HTTPAuthenticate(opts ...ValidationOption) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token and pass it around in the context
			token, err := auth0.FromAuthHeader(r)
			if err != nil {
				log.Error().Err(err).Msg("Failed to extract token from header")
				w.WriteHeader(http.StatusUnauthorized)
				sendFormattedError(w, err, http.StatusUnauthorized, 45, "http.go")
				return
			}

			p, err := HTTPValidateToken(r.Context(), token, opts...)
			if err != nil {
				log.Error().Err(err).Msg("Failed to validate token")
				w.WriteHeader(http.StatusUnauthorized)
				sendFormattedError(w, err, http.StatusUnauthorized, 53, "http.go")
				return
			}

			// Inject token and user permissions into the request context if token is valid
			r = r.WithContext(
				context.WithValue(InjectPermissions(r.Context(), p), jwtKey, token),
			)
			next.ServeHTTP(w, r)
		})
	})
}

// HTTPValidateToken validates the token being passed into the request context.
func HTTPValidateToken(ctx context.Context, token string, opts ...ValidationOption) (*Permissions, error) {
	var options ValidationOptions
	for _, o := range opts {
		o(&options)
	}
	options = *setDefaults(&options)

	p, err := authorize(ctx, options, token)
	if err != nil {
		return p, err
	}
	return p, nil
}
