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
	var proxyExists bool
	var op []byte
	var intEntries []InterfaceEntry
	var err error

	logger.Info().Str("service", cfg.ServiceName).Msg("starting --generate")

	// see if we have already generated the gateway proxy code (<service-name>.pb.gw.go)
	if proxyExists, err = fileExists(cfg.GeneratedPBProxyPath()); err != nil {
		return errors.Wrapf(err, "fileExists(%s)", cfg.GeneratedPBProxyPath())
	}

	apisPath := path.Join(
		cfg.ServicePath(),
		"vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis",
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
			outputDef:  fmt.Sprintf("--grpc-gateway_out=logtostderr=true:%s", cfg.ProtoPath()),
		},
	} {
		logger.Info().Str("service", cfg.ServiceName).
			Str("generating", entry.outputDef).Msg("")
		logger.Debug().Str("service", cfg.ServiceName).Msgf(
			"protoc -I %s --proto_path %s --plugin=%s %s %s",
			apisPath, cfg.ProtoPath(), entry.pluginPath,
			entry.outputDef, cfg.ProtoFilePath(),
		)
		cmd := exec.Command(
			"protoc",
			"-I",
			apisPath,
			"--proto_path",
			cfg.ProtoPath(),
			fmt.Sprintf("--plugin=%s", entry.pluginPath),
			entry.outputDef,
			cfg.ProtoFilePath(),
		)
		if op, err = cmd.CombinedOutput(); err != nil {
			return errors.Wrapf(err, "protoc %v: %s", entry, string(op))
		}
	}

	logger.Info().Msg("parsing generated pb file")
	if intEntries, err = parseGeneratedPBFile(cfg, logger); err != nil {
		return errors.Wrap(err, "parseGeneratedPBFile")
	}

	logger.Info().Msgf("interface defines %d methods", len(intEntries))

	if err = createMethodFiles(cfg, logger, intEntries); err != nil {
		return errors.Wrap(err, "createMethodFiles")
	}

	// if this is the first time we have generated the gateway proxy code
	// write the real proxy template over the stub we stored in --init
	if !proxyExists {
		if proxyExists, err = fileExists(cfg.GeneratedPBProxyPath()); err != nil {
			return errors.Wrapf(err, "fileExists(%s)", cfg.GeneratedPBProxyPath())
		}
		if proxyExists {
			if err = writeProxyTemplate(cfg, logger); err != nil {
				return errors.Wrap(err, "writeProxyTemplate")
			}
		}
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

func createMethodFiles(
	cfg config.Config,
	logger zerolog.Logger,
	intEntries []InterfaceEntry,
) error {
	var err error

METHOD_LOOP:
	for _, entry := range intEntries {
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

func writeProxyTemplate(cfg config.Config, logger zerolog.Logger) error {
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
			ServiceName   string
			GoServiceName string
			PBImport      string
		}{
			cfg.ServiceName,
			cfg.GoServiceName(),
			cfg.PBImportPath(),
		},
	)

	if err != nil {
		return errors.Wrap(err, "templ.Merge")
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

	cmd := exec.Command("dep", "ensure", "-v")
	if op, err = cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "dep ensure; '%s'", string(op))
	}

	return nil
}
