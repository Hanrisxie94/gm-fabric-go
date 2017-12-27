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

package initsvc

var oauthTemplate = `package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/deciphernow/gm-fabric-go/oauth"
)

func putOauthInCtxIfNeeded(ctx context.Context) context.Context {
	if viper.GetBool("use_oauth") {
		return oauth.ContextWithOptions(
			ctx,
			oauth.WithProvider(viper.GetString("oauth_provider")),
			oauth.WithClientID(viper.GetString("oauth_client_id")),
		)
	}
	return ctx
}

func getOauthOptsIfNeeded(logger zerolog.Logger) ([]grpc.ServerOption, error) {
	var err error

	if !viper.GetBool("use_oauth") {
		return nil, nil
	}

	provider := viper.GetString("oauth_provider")
	clientID := viper.GetString("oauth_client_id")

	logger.Debug().Str("service", "test_service").
		Str("oauth_provider", provider).
		Str("oauth_client_id", clientID).
		Msg("loading OAuth config")

	interceptor, err := oauth.NewOauthInterceptor(
		oauth.WithProvider(provider),
		oauth.WithClientID(clientID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "oauth.NewOauthInterceptor")
	}

	return []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(interceptor)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(interceptor)),
	}, nil
}

`
