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
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templ"
)

// InitService initializes the service up to the point where it needs a
// protocol buffer definiton
func InitService(cfg config.Config, logger zerolog.Logger) error {
	var err error

	logger.Info().Str("service", cfg.ServiceName).Msg("starting --init")

	// TODO: build in a temp dir and rename on success - no partial completion

	logger.Info().Msg("creating directories")

	for _, dirPath := range []string{
		cfg.ServicePath(),
		cfg.CmdPath(),
		cfg.ServerPath(),
		cfg.ConfigPackagePath(),
		cfg.MethodsPath(),
		cfg.GRPCClientPath(),
		cfg.DockerPath(),
		cfg.ProtoPath(),
	} {
		if err = os.Mkdir(dirPath, 0777); err != nil {
			return errors.Wrapf(err, "os.Mkdir(%s)", dirPath)
		}
	}

	logger.Info().Msg("creating server")
	if err = createServer(cfg, logger); err != nil {
		return errors.Wrap(err, "createServer")
	}

	logger.Info().Msg("creating grpc client")
	if err = createGRPCClient(cfg, logger); err != nil {
		return errors.Wrap(err, "createGRPCClient")
	}

	logger.Info().Msg("creating docker files")
	if err = createDocker(cfg, logger); err != nil {
		return errors.Wrap(err, "createDocker")
	}

	logger.Info().Msgf("creating %s", cfg.ProtoFileName())
	err = templ.Merge(
		"proto",
		protoTemplate,
		cfg.ProtoFilePath(),
		struct {
			ProtoDirName  string
			GoServiceName string
		}{
			cfg.ProtoDirName(),
			// 2017-10-23 dougfort -- using GoServiceName here instead of ProtoServiceName
			// because, while the grpc plugin converts the servicename, the gateway plugin
			// does not. So if we put the coverted name here, they both come up with the
			// same name.
			cfg.GoServiceName(),
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.ProtoFileName())
	}

	logger.Info().Msgf("creating %s", cfg.BuildServerScriptName())
	err = templ.Merge(
		"build",
		buildServerTemplate,
		cfg.BuildServerScriptPath(),
		struct {
			ServerPath  string
			ServiceName string
		}{
			cfg.ServerPath(),
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.BuildServerScriptName())
	}
	if err = os.Chmod(cfg.BuildServerScriptPath(), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", cfg.BuildServerScriptPath())
	}

	logger.Info().Msgf("creating %s", cfg.BuildGRPCClientScriptName())
	err = templ.Merge(
		"build-client",
		buildGRPCClientTemplate,
		cfg.BuildGRPCClientScriptPath(),
		struct {
			GRPCClientPath string
			ServiceName    string
		}{
			cfg.GRPCClientPath(),
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.BuildGRPCClientScriptName())
	}
	if err = os.Chmod(cfg.BuildGRPCClientScriptPath(), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", cfg.BuildGRPCClientScriptPath())
	}

	logger.Info().Msgf("creating %s", cfg.BuildDockerImageScriptName())
	err = templ.Merge(
		"build-docker-image",
		buildDockerImageTemplate,
		cfg.BuildDockerImageScriptPath(),
		struct {
			ServiceName      string
			ServerPath       string
			DockerPath       string
			SettingsFilePath string
		}{
			cfg.ServiceName,
			cfg.ServerPath(),
			cfg.DockerPath(),
			cfg.SettingsFilePath(),
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.BuildDockerImageScriptName())
	}
	if err = os.Chmod(cfg.BuildDockerImageScriptPath(), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", cfg.BuildDockerImageScriptPath())
	}

	logger.Info().Msgf("creating %s", cfg.RunDockerImageScriptName())
	err = templ.Merge(
		"build-run-docker-script",
		runDockerImageTemplate,
		cfg.RunDockerImageScriptPath(),
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
		return errors.Wrapf(err, "creating %s", cfg.RunDockerImageScriptName())
	}
	if err = os.Chmod(cfg.RunDockerImageScriptPath(), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", cfg.RunDockerImageScriptPath())
	}

	logger.Info().Msg("creating a .gitignore")
	err = templ.Merge(
		"gitignore",
		gitIgnoreTemplate,
		cfg.GitIgnorePath(),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "creating .gitignore")
	}

	err = BundleRPMArtifacts(cfg, logger)
	if err != nil {
		return errors.Wrap(err, "creating rpm artifacts")
	}

	logger.Info().Msg("initializing versioning")
	if err = initVersioning(cfg, logger); err != nil {
		return errors.Wrap(err, "initVersioning()")
	}

	logger.Info().Msg("--init complete")
	return nil
}
