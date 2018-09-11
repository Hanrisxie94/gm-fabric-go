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
package listauth

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deciphernow/gm-fabric-go/listauth/auth"
	"github.com/rs/zerolog"
)

func TestHTTP(t *testing.T) {
	// testData adapted from auth/auth_test.go
	type testData struct {
		name           string
		dn             string
		s              string
		expectedStatus int
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	for _, td := range []testData{
		// with both lists empty, we'll take anything
		{
			name: "empty lists",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": [],
				"userWhitelist": []
			}`,
			expectedStatus: http.StatusOK,
		},
		// if the whitelist applies to the dn, it is authorized
		{
			name: "whitelist only",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			expectedStatus: http.StatusOK,
		},
		// if the blacklist applies to the dn, it is unauthorized
		{
			name: "blacklist only",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": ["cn=alec.holmes,dc=deciphernow,dc=com"],
				"userWhitelist": []
			}`,
			expectedStatus: http.StatusUnauthorized,
		},
		// if the blacklist applies to the dn, it is unauthorized
		// even if it is on the whitelist
		{
			name: "blacklist and whitelist",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": ["cn=alec.holmes,dc=deciphernow,dc=com"],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			expectedStatus: http.StatusUnauthorized,
		},
		// the whitelist does not apply to the dn, if the 'dc' RDNs are
		// in a different order
		{
			name: "RDNs out of order",
			dn:   "cn=alec.holmes,dc=com,dc=deciphernow",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			expectedStatus: http.StatusUnauthorized,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
				Output(zerolog.ConsoleWriter{Out: os.Stderr})
			a, err := auth.New(strings.NewReader(td.s))
			if err != nil {
				t.Fatalf("%s: auth.New failed: %v", td.name, err)
			}
			mwFunc := HTTPAuthenticate(logger, a)
			handler := mwFunc(okHandler{})

			sw := statusWriter{}

			// TODO: to get full coverage, we need to test both certificates
			// and headers
			cert := x509.Certificate{}

			req := http.Request{
				Header: make(http.Header),
				TLS: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{&cert},
				},
			}

			req.Header.Add("USER_DN", td.dn)

			handler.ServeHTTP(&sw, &req)

			if sw.statusCode != td.expectedStatus {
				t.Fatalf("status %d does not match expectd status %d",
					sw.statusCode, td.expectedStatus)
			}
		})
	}
}

// okHandler always returns Ok
type okHandler struct {
}

func (ok okHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// statusWriter is an http.ResponseWriter that captures the returning status
type statusWriter struct {
	statusCode int
}

// Header returns the header map that will be sent by
// WriteHeader.
func (w *statusWriter) Header() http.Header {
	return make(http.Header)
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *statusWriter) Write([]byte) (int, error) {
	return 0, nil
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (w *statusWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
