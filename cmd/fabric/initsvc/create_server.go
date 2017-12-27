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
	"path/filepath"

	"github.com/rs/zerolog"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templ"
)

func createServer(
	cfg config.Config,
	logger zerolog.Logger,
) error {
	var err error

	logger.Info().Msg("creating server main.go")
	err = templ.Merge(
		"server",
		mainTemplate,
		filepath.Join(cfg.ServerPath(), "main.go"),
		struct {
			ServiceName       string
			GoServiceName     string
			ConfigPackage     string
			ConfigPackageName string
			MethodsPackage    string
			PBImport          string
		}{
			cfg.ServiceName,
			cfg.GoServiceName(),
			cfg.ConfigPackageImportPath(),
			cfg.ConfigPackageName(),
			cfg.MethodsImportPath(),
			cfg.PBImportPath(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating server main.go")
	}
	err = templ.Merge(
		"tls",
		tlsTemplate,
		filepath.Join(cfg.ServerPath(), "tls.go"),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating server tls.go")
	}
	err = templ.Merge(
		"oauth",
		oauthTemplate,
		filepath.Join(cfg.ServerPath(), "oauth.go"),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating server oauth.go")
	}
	err = templ.Merge(
		"statsd",
		statsdTemplate,
		filepath.Join(cfg.ServerPath(), "statsd.go"),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating server oauth.go")
	}
	err = templ.Merge(
		"zk",
		zkTemplate,
		filepath.Join(cfg.ServerPath(), "zk.go"),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating server oauth.go")
	}

	// The gateway proxy code can't compile until the developer generates
	// HTTP option methods from the .proto file. That may never happen.
	// So here we write out a stub to keep the compiler happy.
	// In the --generate phase, if we detect HTTP Option methods,
	// we will write over this stub with the real gateway proxy code.
	logger.Info().Msg("creating server gateway_proxy.go (stub)")
	err = writeProxyStub(filepath.Join(cfg.ServerGatewayProxySourceFilePath()))
	if err != nil {
		return errors.Wrap(err, "creating server gateway_proxy.go (stub)")
	}

	logger.Info().Msgf("creating %s", cfg.ConfigPackageName())
	err = templ.Merge(
		"config",
		configPackageTemplate,
		filepath.Join(cfg.ConfigPackagePath(), "config.go"),
		struct {
			ServiceName       string
			ConfigPackageName string
		}{
			cfg.ServiceName,
			cfg.ConfigPackageName(),
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.ConfigPackageName())
	}

	logger.Info().Msgf("creating %s", cfg.MethodsPackageName())
	err = templ.Merge(
		"methods",
		methodsTemplate,
		filepath.Join(cfg.MethodsPath(), "new_server.go"),
		struct {
			MethodsPackageName string
			ConfigPackage      string
			ConfigPackageName  string
			PBImport           string
			GoServiceName      string
		}{
			cfg.MethodsPackageName(),
			cfg.ConfigPackageImportPath(),
			cfg.ConfigPackageName(),
			cfg.PBImportPath(),
			cfg.GoServiceName(),
		},
	)
	if err != nil {
		return errors.Wrapf(err, "creating %s", cfg.MethodsPackageName())
	}

	return nil
}

func writeProxyStub(proxySourceFilePath string) error {
	var file *os.File
	var err error

	if file, err = os.Create(proxySourceFilePath); err != nil {
		return errors.Wrapf(err, "os.Create(%s)", proxySourceFilePath)
	}
	defer file.Close() // ignoring error return

	if _, err = file.WriteString(proxyStubTemplate); err != nil {
		return errors.Wrapf(err, "file.WriteString")
	}

	return nil
}
