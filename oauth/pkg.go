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

// Package oauth provides HTTP and GRPC middleware that supports HS256 (HMAC) token validation and verification.
package oauth

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

type key int

// SignAlg is a key to deal with token algorithm signing methods.
// We use a custom type here to regulate strings.
type SignAlg string

const (
	permissionsKey key = iota
	jwtKey
	optionsKey
	// RS256 ...
	RS256 SignAlg = "RS256"
	// HS256 ...
	HS256 SignAlg = "HS256"
)

func init() {
	// Change zerolog output to Stdout instead of Stderr
	log.Output(os.Stdout)
}

// ValidationOption follows functional opts pattern
type ValidationOption func(*ValidationOptions)

// ValidationOptions follows functional opts pattern
type ValidationOptions struct {
	HMACSecret string  `json:"hmac_secret"`
	RSAKeyPath string  `json:"rsa_key_path"`
	SigningAlg SignAlg `json:"signing_algorithm"`
	ClientID   string  `json:"dex_client_id"`
	Provider   string  `json:"oauth_provider"`
}

// WithHMACSecret will pass the token secret to the validation function if the user has selected HMAC (HSA)
func WithHMACSecret(secret string) ValidationOption {
	return func(o *ValidationOptions) {
		o.HMACSecret = secret
	}
}

// WithRSAKeyPath will pass the file path of a public to the validation function if the user has selected RSA
func WithRSAKeyPath(keyPath string) ValidationOption {
	return func(o *ValidationOptions) {
		o.RSAKeyPath = keyPath
	}
}

// WithSigningAlg will pass the signing algorithm to the validation function
func WithSigningAlg(alg SignAlg) ValidationOption {
	return func(o *ValidationOptions) {
		o.SigningAlg = alg
	}
}

// WithClientID will pass the dex client id to the validation function
func WithClientID(id string) ValidationOption {
	return func(o *ValidationOptions) {
		o.ClientID = id
	}
}

// WithProvider will pass the provider url to the validation function
func WithProvider(provider string) ValidationOption {
	return func(o *ValidationOptions) {
		o.Provider = provider
	}
}

// Permissions holds the basic viewer authorities.
// Should be returned by our OAuth server.
// Not sure what to put in here yet but it should conform to what we decide to do in the OAuth server.
type Permissions struct {
	Groups   []string `json:"groups"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Verified bool     `json:"email_verified"`
}

// InjectPermissions will add a user permission object into a context
func InjectPermissions(ctx context.Context, p *Permissions) context.Context {
	return context.WithValue(ctx, permissionsKey, p)
}

// RetrievePermissions will extract a users permissions object from a context.
func RetrievePermissions(ctx context.Context) *Permissions {
	p, ok := ctx.Value(permissionsKey).(*Permissions)
	if !ok {
		return nil
	}

	return p
}

// LoadUserPermissions will read a reader and decode the response into a Permissions object.
func LoadUserPermissions(r io.Reader) (*Permissions, error) {
	var p *Permissions

	err := json.NewDecoder(r).Decode(&p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// RetrieveToken will extract the JWT from the request context if a user wishes to do other processing.
func RetrieveToken(ctx context.Context) string {
	token, ok := ctx.Value(jwtKey).(string)
	if !ok {
		return ""
	}
	return token
}
