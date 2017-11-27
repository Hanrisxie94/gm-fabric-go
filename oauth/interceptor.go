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

/*
	Example code usage:

	ctx := oauth.ContextWithOptions(oauth.WithSigningAlg(HS256), oauth.WithTokenSecret("KbtfnXOHH3ezniXIsHYSd4WhZPBXH1vB"))

	server := grpc.NewServer(
	    grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(oauth.ValidateToken(ctx))),
	    grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(oauth.ValidateToken(ctx))),
	)

	return server



	Other example that could serve as a specific handler auth check:

	// TODO: Provide server opts functions to pass around necessary data such as key and signing alg


	func (s *Server) GetItems(ctx context.Context, in *store.Item) (*store.Items, error) {
	    // Auth check
	    _, err := oauth.ValidateToken(ctx)
	    if err != nil {
	        return nil, err
	    }

	    // If token validation was successful, get user permissions
	    permissions := oauth.RetrievePermissionsFromContext(ctx)
	    if permissions == nil {
	        return nil, errors.New("user permissions can not be nil")
	    }

	    // Do logic
	    // Only return items a user is allowed to see

	    return items, nil
	}
*/

package oauth

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// ValidateToken validates the token being passed into the request metadata.
// A developer should pass the tokenSecret in the context with the key: 'token-secret'
func ValidateToken(ctx context.Context) (context.Context, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, errors.New("Failed to retrieve rpc metadata from incoming context")
	}

	opts, err := OptionsFromIncomingContext(ctx)
	if !ok {
		return ctx, err
	}

	p, err := ValidateOnSigningMethod(ctx, meta, opts...)
	if err != nil {
		log.Error().Err(err).Msg("Interceptor failed to validate token")
		return ctx, err
	}

	ctx = InjectPermissions(ctx, p)

	return ctx, nil
}

// ValidateOnSigningMethod will extract a signing algorithm type from the metadata context and validate accordingly.
func ValidateOnSigningMethod(ctx context.Context, meta metadata.MD, opts ...ValidationOption) (*Permissions, error) {
	var options ValidationOptions
	for _, o := range opts {
		o(&options)
	}

	token, err := ExtractTokenFromMD(meta)
	if err != nil {
		return nil, err
	}

	p, err := authorize(ctx, options, token)
	if err != nil {
		return p, err
	}
	return p, nil
}

// ContextWithOptions follows the functional opts pattern, but injects the opts into the context to follow the gRPC interceptor pattern
func ContextWithOptions(ctx context.Context, opts ...ValidationOption) context.Context {
	return context.WithValue(ctx, optionsKey, opts)
}

// OptionsFromIncomingContext will retrieve the injected opts from the context
func OptionsFromIncomingContext(ctx context.Context) ([]ValidationOption, error) {
	options, ok := ctx.Value(optionsKey).([]ValidationOption)
	if !ok {
		return nil, errors.New("Failed to retrieve validation options from context")
	}

	return options, nil
}

// ExtractTokenFromMD will pull the bearer token from the headers passed down.
// Probably should make this more robust in the future as it only takes very specific input.
// Format must be:
// Authorization: Bearer $TOKEN
func ExtractTokenFromMD(meta metadata.MD) (string, error) {
	if len(meta["Authorization"]) == 0 {
		return "", errors.New("Missing authorization header")
	}

	// Also only handle a single token which is why we get the first position
	bearer := strings.Split(meta["Authorization"][0], " ")
	// Token lives in the 2nd position since the 1st will be the value: "Bearer"
	token := bearer[1]
	if token == "" {
		return token, errors.New("Token is empty")
	}

	return token, nil
}
