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

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/rs/zerolog"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templates"
)

// GenerateProtobuf generates code from protocol buffer definitions
func GenerateProtobuf(cfg config.Config, logger zerolog.Logger) error {
	var op []byte
	var serverDef ServerInterfaceData
	var err error

	logger.Info().Str("service", cfg.ServiceName).Msg("starting --generate")

	apisPath := path.Join(
		cfg.ServicePath(),
		"vendor",
		"github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis",
	)

	for _, entry := range []struct {
		pluginPath string
		outputDef  string
	}{
		{
			pluginPath: cfg.ProtocGenGoPluginPath(),
			outputDef:  fmt.Sprintf("--go_out=plugins=grpc:%s", cfg.ProtoPath()),
		},
		{
			pluginPath: cfg.ProtocGenGatewayPluginPath(),
			outputDef:  fmt.Sprintf("--grpc-gateway_out=logtostderr=true,allow_delete_body=true:%s", cfg.ProtoPath()),
		},
		{
			pluginPath: cfg.ProtocGenSwaggerPluginPath(),
			outputDef:  fmt.Sprintf("--swagger_out=logtostderr=true,allow_delete_body=true:%s", cfg.ProtoPath()),
		},
	} {
		logger.Info().Str("service", cfg.ServiceName).
			Str("generating", entry.outputDef).Msg("")
		cmd := exec.Command("protoc")
		cmd.Args = append(cmd.Args, "-I")
		cmd.Args = append(cmd.Args, apisPath)
		for _, protocInclude := range cfg.ProtocIncludePaths {
			cmd.Args = append(cmd.Args, "-I")
			cmd.Args = append(cmd.Args, protocInclude)
		}
		cmd.Args = append(cmd.Args, fmt.Sprintf("--proto_path=%s", cfg.ProtoPath()))
		cmd.Args = append(cmd.Args, fmt.Sprintf("--plugin=%s", entry.pluginPath))
		cmd.Args = append(cmd.Args, entry.outputDef)
		cmd.Args = append(cmd.Args, cfg.ProtoFilePath())
		logger.Debug().Str("service", cfg.ServiceName).Msgf("protoc %s", cmd.Args)
		if op, err = cmd.CombinedOutput(); err != nil {
			return errors.Wrapf(err, "protoc %v: %s", entry, string(op))
		}
	}

	logger.Info().Msg("parsing generated pb file")
	if serverDef, err = parseGeneratedPBFile(cfg, logger); err != nil {
		return errors.Wrap(err, "parseGeneratedPBFile")
	}

	logger.Info().Msgf("interface (%sServer) defines %d methods",
		serverDef.ServerName, len(serverDef.Prototypes))

	if err = initializeMethodsDir(cfg, logger, serverDef); err != nil {
		return errors.Wrap(err, "initializeMethodsDir")
	}

	if err = createMethodFiles(cfg, logger, serverDef); err != nil {
		return errors.Wrap(err, "createMethodFiles")
	}

	logger.Info().Str("service", cfg.ServiceName).Msg("--generate complete")
	return nil
}

func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "os.Stat(%s)", filePath)
	}
	return true, nil
}

func initializeMethodsDir(
	cfg config.Config,
	logger zerolog.Logger,
	serverDef ServerInterfaceData,
) error {
	var methodsExists bool
	var err error

	methodsExists, err = fileExists(cfg.MethodsPath())
	if err != nil {
		return errors.Wrapf(err, "fileExists(%s)", cfg.MethodsPath())
	}

	if !methodsExists {
		if err = os.Mkdir(cfg.MethodsPath(), os.ModePerm); err != nil {
			return errors.Wrapf(err, "os.Mkdir(%s)", cfg.MethodsPath())
		}
	}

	// See issue #312
	// The methods directory may be in an indeterminate state due to some
	// earlier failure. So let's write out the template files each time.
	// It doesn't cost very much and it leaves us in a known state.

	if err = writeServerNew(cfg, logger, serverDef); err != nil {
		return errors.Wrap(err, "writeServerNew")
	}

	return writeProxyTemplate(cfg, logger, serverDef)
}

func createMethodFiles(
	cfg config.Config,
	logger zerolog.Logger,
	serverDef ServerInterfaceData,
) error {
	var err error

METHOD_LOOP:
	for _, entry := range serverDef.Prototypes {
		var exists bool

		methodFileName := computeMethodFileName(entry.Prototype)
		methodFilePath := path.Join(cfg.MethodsPath(), methodFileName)

		if exists, err = fileExists(methodFilePath); err != nil {
			return errors.Wrapf(err, "fileExsts(%s)", methodFilePath)
		}
		if exists {
			logger.Debug().Msgf("method file %s already exists, ignoring",
				methodFileName)
			continue METHOD_LOOP
		}

		if err = generateMethod(cfg, logger, entry, methodFilePath); err != nil {
			return errors.Wrap(err, "generateMethod")
		}
	}

	return nil
}

func writeServerNew(
	cfg config.Config,
	logger zerolog.Logger,
	serverDef ServerInterfaceData,
) error {
	var templ string
	var err error

	templ, err = loadTemplateFromCache(cfg, logger, "new_server.go")
	if err != nil {
		return errors.Wrapf(err, "loadTemplateFromCache %s", "new_server.go")
	}

	newServerFilePath := path.Join(cfg.MethodsPath(), "new_server.go")

	err = templates.Merge(
		"newserver",
		templ,
		newServerFilePath,
		struct {
			MethodsPackageName  string
			PBImport            string
			GoServiceName       string
			ServerInterfaceName string
		}{
			cfg.MethodsPackageName(),
			cfg.PBImportPath(),
			cfg.GoServiceName(),
			serverDef.ServerName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "templ.Merge")
	}

	cmd := exec.Command(
		"gofmt",
		"-w",
		newServerFilePath,
	)
	if op, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "gofmt -w %s: %s",
			newServerFilePath, string(op))
	}

	return nil
}

func writeProxyTemplate(
	cfg config.Config,
	logger zerolog.Logger,
	serverDef ServerInterfaceData,
) error {
	var op []byte
	var cwd string
	var templ string
	var err error

	templ, err = loadTemplateFromCache(cfg, logger, "gateway_proxy.go")
	if err != nil {
		return errors.Wrapf(err, "loadTemplateFromCache %s", "gateway_proxy.go")
	}

	err = templates.Merge(
		"proxy",
		templ,
		cfg.ServerGatewayProxySourceFilePath(),
		struct {
			ServiceName         string
			GoServiceName       string
			PBImport            string
			ServerInterfaceName string
		}{
			cfg.ServiceName,
			cfg.GoServiceName(),
			cfg.PBImportPath(),
			serverDef.ServerName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "templ.Merge")
	}

	cmd := exec.Command(
		"gofmt",
		"-w",
		cfg.ServerGatewayProxySourceFilePath(),
	)
	if op, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "gofmt -w %s: %s",
			cfg.ServerGatewayProxySourceFilePath(), string(op))
	}

	// now we have some new dependencies
	if cwd, err = os.Getwd(); err != nil {
		return errors.Wrap(err, "os.Getwd()")
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			panic(err)
		}
	}()
	if err = os.Chdir(cfg.ServicePath()); err != nil {
		return errors.Wrapf(err, "os.Chdir(%s)", cfg.ServicePath())
	}

	cmd = exec.Command("dep", "ensure", "-v")
	if op, err = cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "dep ensure; '%s'", string(op))
	}

	return nil
}
