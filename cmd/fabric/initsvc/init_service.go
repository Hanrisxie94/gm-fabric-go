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
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templates"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// InitService initializes the service up to the point where it needs a
// protocol buffer definiton
func InitService(cfg config.Config, logger zerolog.Logger) error {

	var err error
	var data interface{}

	data = struct {
		ServiceName        string
		ServicePath        string
		GoServiceName      string
		ProtoServiceName   string
		ConfigPackage      string
		ConfigPackageName  string
		MethodsPackage     string
		MethodsPackageName string
		ProtoDirName       string
		PBImport           string
	}{
		cfg.ServiceName,
		cfg.ServicePath(),
		cfg.GoServiceName(),
		cfg.ProtoServiceName(),
		cfg.ConfigPackageImportPath(),
		cfg.ConfigPackageName(),
		cfg.MethodsImportPath(),
		cfg.MethodsPackageName(),
		cfg.ProtoDirName(),
		cfg.PBImportPath(),
	}

	logger.Info().Str("service", cfg.ServiceName).Msg("starting --init")

	var output []byte
	var callback = func(name *template.Template, content *template.Template, mode os.FileMode) error {
		if err = templates.Render(name, content, mode, cfg.ServicePath(), data, logger); err != nil {
			return errors.Wrapf(err, "Failed to render template for %s", name.Name())
		}
		return nil
	}

	logger.Debug().Msg(fmt.Sprintf("Fetching template from %s", cfg.TemplateUrl))

	if err = templates.Fetch(cfg.TemplateUrl, callback, logger); err != nil {
		return errors.Wrapf(err, "Failed to fetch template from %s", cfg.TemplateUrl)
	}

	if err = within(cfg.ServicePath(), func() error {
		// assume we got a template for Gopkg.toml
		if output, err = exec.Command("dep", "ensure").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	if err = within(path.Join(cfg.VendorPath(), "github.com", "golang", "protobuf", cfg.ProtocGenGoPluginName()), func() error {
		logger.Debug().Msg("Installing Golang Plugin...")
		if output, err = exec.Command("go", "install", "-v").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	if err = within(path.Join(cfg.VendorPath(), "github.com", "grpc-ecosystem", "grpc-gateway", cfg.ProtocGenGatewayPluginName()), func() error {
		logger.Debug().Msg("Installing Gateway Plugin...")
		if output, err = exec.Command("go", "install", "-v").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	return nil

}

func within(directory string, callback func() error) error {

	var err error
	var current string

	if current, err = os.Getwd(); err != nil {
		return errors.Wrap(err, "Failed discover current working directory")
	}

	if err = os.Chdir(directory); err != nil {
		return errors.Wrapf(err, "Failed to change working directory to %s", directory)
	}

	defer func() {
		if err = os.Chdir(current); err != nil {
			panic(err)
		}
	}()

	if err = callback(); err != nil {
		return errors.Wrapf(err, "Failed executing callback from working directory %s", directory)
	}

	return nil
}
