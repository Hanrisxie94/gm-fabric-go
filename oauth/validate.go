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
	"errors"
	"fmt"
	"io/ioutil"

	oidc "github.com/coreos/go-oidc"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
)

// validate will parse a token, verify and return wether the token is valid or not.
func validateHS256(token, secret string) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}

	// Perform series of checks to validate token
	if jwt.SigningMethodHS256.Alg() != parsedToken.Header["alg"] {
		message := fmt.Sprintf("Expected %s signing method but token specified %s",
			jwt.SigningMethodHS256.Alg(),
			parsedToken.Header["alg"])
		err := fmt.Errorf("Error validating token algorithm: %s", message)
		log.Error().Err(err).Msg("Error validating token algorithm")
		return err
	}

	// If the token was invalid originally, fail the validation
	if !parsedToken.Valid {
		return errors.New("Token is invalid")
	}

	return nil
}

// validate will parse a token, verify and return wether the token is valid or not
// Follows RS265 spec - requires a cert and key
func validateRS256(token, keyPath string) error {
	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			return nil, err
		}

		return key, nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse token")
		return err
	}

	if jwt.SigningMethodRS256.Alg() != parsedToken.Header["alg"] {
		message := fmt.Sprintf("Expected %s signing method but token specified %s",
			jwt.SigningMethodRS256.Alg(),
			parsedToken.Header["alg"])
		err := fmt.Errorf("Error validating token algorithm: %s", message)
		log.Error().Err(err).Msg("Error validating token algorithm")
		return err
	}

	if !parsedToken.Valid {
		return errors.New("Token is invalid")
	}

	return nil
}

func validateWithDex(ctx context.Context, token string, options ValidationOptions) (*Permissions, error) {
	provider, err := oidc.NewProvider(ctx, options.Provider)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new provider")
		return nil, err
	}

	idTokenVerifier := provider.Verifier(
		&oidc.Config{
			ClientID: options.ClientID,
		},
	)

	// Parse and verify ID Token payload.
	idToken, err := idTokenVerifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	// Extract custom claims.
	var p Permissions
	if err := idToken.Claims(&p); err != nil {
		return nil, fmt.Errorf("Failed to parse claims: %v", err)
	}
	if !p.Verified {
		return nil, fmt.Errorf("Email (%q) in returned claims was not verified", p.Email)
	}

	return &p, nil
}

func authorize(ctx context.Context, options ValidationOptions, token string) (*Permissions, error) {
	var err error

	// Block the authorization if no token is sent
	if token == "" {
		return nil, fmt.Errorf("No token was sent, cancelling authorization")
	}

	// Authenticate based on the signing method since each type of validation is different
	switch {
	case options.SigningAlg == RS256:
		err = validateRS256(token, options.RSAKeyPath)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case options.SigningAlg == HS256:
		if options.HMACSecret == "" {
			return nil, errors.New("missing token secret")
		}
		err = validateHS256(token, options.HMACSecret)
		if err != nil {
			return nil, err
		}
		return nil, nil
	default:
		p, err := validateWithDex(ctx, token, options)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
}
