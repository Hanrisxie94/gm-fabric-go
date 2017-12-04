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

import (
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/cmd/gm-servgen/config"
	"github.com/deciphernow/gm-fabric-go/cmd/gm-servgen/templ"
)

func createDocker(
	cfg config.Config,
	logger zerolog.Logger,
) error {
	var err error

	logger.Info().Msg("creating Dockerfile")
	err = templ.Merge(
		"Dockerfile",
		dockerFileTemplate,
		filepath.Join(cfg.DockerPath(), "Dockerfile"),
		struct {
			ServiceName       string
			GrpcServerPort    int
			MetricsServerPort int
			GatewayProxyPort  int
		}{
			cfg.ServiceName,
			viper.GetInt("grpc_server_port"),
			viper.GetInt("metrics_server_port"),
			viper.GetInt("gateway_proxy_port"),
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating Dockerfile")
	}

	logger.Info().Msg("creating docker entry point")
	err = templ.Merge(
		"entrypoint",
		entryPointTemplate,
		cfg.DockerEntryPointPath(),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating docker entry point")
	}

	return nil
}
