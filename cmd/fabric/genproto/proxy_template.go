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

package genproto

var proxyTemplate = `package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"

	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/deciphernow/gm-fabric-go/middleware"

	pb "{{.PBImport}}"
)

func gatewayProxy(
	ctx context.Context,
	logger zerolog.Logger,
) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.Register{{.GoServiceName}}HandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf(
			"%s:%d",
			viper.GetString("grpc_server_host"),
			viper.GetInt("grpc_server_port"),
		),
		opts,
	)
	if err != nil {
		logger.Fatal().AnErr("pb.Register{{.GoServiceName}}HandlerFromEndpoint", err).Msg("")
	}

	logger.Debug().Str("service", "{{.ServiceName}}").
		Str("host", viper.GetString("gateway_proxy_host")).
		Int("port", viper.GetInt("gateway_proxy_port")).
		Msg("starting gateway proxy server")

	if viper.GetBool("verbose_logging") {
		stack := middleware.Chain(
			middleware.MiddlewareFunc(hlog.NewHandler(logger)),
			middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
				hlog.FromRequest(r).Info().
					Str("method", r.Method).
					Str("path", r.URL.String()).
					Int("status", status).
					Int("size", size).
					Dur("duration", duration).
					Msg("Access")
			})),
			middleware.MiddlewareFunc(hlog.UserAgentHandler("user_agent")),
		)

		// http.ListenAndServe blocks until cancelled. It always returns a non-nil error
		// We're checking here so we don't lose an error at startup
		err = http.ListenAndServe(
			fmt.Sprintf(
				"%s:%d",
				viper.GetString("gateway_proxy_host"),
				viper.GetInt("gateway_proxy_port"),
			),
			stack.Wrap(mux),
		)
		if err != nil {
			logger.Error().AnErr("http.ListenAndServe", err).Msg("")
		}
	} else {
		// http.ListenAndServe blocks until cancelled. It always returns a non-nil error
		// We're checking here so we don't lose an error at startup
		err = http.ListenAndServe(
			fmt.Sprintf(
				"%s:%d",
				viper.GetString("gateway_proxy_host"),
				viper.GetInt("gateway_proxy_port"),
			),
			mux,
		)
		if err != nil {
			logger.Error().AnErr("http.ListenAndServe", err).Msg("")
		}
	}
}
`
