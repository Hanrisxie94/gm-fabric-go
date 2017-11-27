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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
)

func TestWithToken(t *testing.T) {
	s := buildHTTPServer()
	defer s.Close()
	client := s.Client()
	url := s.URL

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Use this token:
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFsZWMuaG9sbWVzQGRlY2lwaGVybm93LmNvbSIsIm5hbWUiOiJBbGVjIEhvbG1lcyJ9.pU6UVjOG2ipw4S-eCN-MZLN8DHwKV8TO0ehKpk81MEo"
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.Body == nil {
		t.Fatal("response was nil")
	}

	var response struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Decode failed")
	}

	// Compare tokens to make sure we authenticated properly
	// If authentication fails, no token will be passed
	if response.Token != token {
		t.Fatalf("tokens did not match")
	}

	fmt.Println("PASS \nok\tSuccessfuly compared tokens")
}

func buildHTTPServer() *httptest.Server {
	// Initiate router
	mux := mux.NewRouter()
	mux.HandleFunc("/", testJWT)

	// Inject the JWT middleware
	stack := middleware.Chain(
		// Use a random 32 bit secret
		HTTPAuthenticate(WithSigningAlg(HS256), WithHMACSecret("KbtfnXOHH3ezniXIsHYSd4WhZPBXH1vB")),
	)

	// Create basic http server
	return httptest.NewServer(stack.Wrap(mux))
}

func testJWT(w http.ResponseWriter, r *http.Request) {
	token := RetrieveToken(r.Context())
	if token == "" {
		log.Error().Msg("Missing JWT")
	}

	var response struct {
		Token string `json:"token"`
	}
	response.Token = token

	// Send the token back for verification purposes
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error().Err(err)
	}
}
