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
	"path"
	"path/filepath"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templ"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// BundleRPMArtifacts will bundle all RPM artifacts in the proper directory needed to deploy with yum
func BundleRPMArtifacts(cfg config.Config, logger zerolog.Logger) error {
	logger.Info().Msg("creating rpm dir")
	err := os.Mkdir(cfg.RPMBundlingPath(), 0777)
	if err != nil {
		return errors.Wrap(err, "creating rpm dir")
	}

	logger.Info().Msg("creating rpm build artifacts")
	err = templ.Merge(
		"build-rpm-build",
		buildRPMTemplate,
		filepath.Join(cfg.RPMBundlingPath(), "build"),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "creating rpm/build")
	}
	if err = os.Chmod(path.Join(cfg.RPMBundlingPath(), "build"), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", path.Join(cfg.RPMBundlingPath(), "build"))
	}

	logger.Info().Msg("creating rpm bundle script")
	err = templ.Merge(
		"build-rpm-bundling",
		buildRPMScriptTemplate,
		filepath.Join(cfg.RPMBundlingPath(), "script"),
		struct {
			ServiceName string
			ServicePath string
		}{
			cfg.ServiceName,
			cfg.ServicePath(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating rpm/script")
	}
	if err = os.Chmod(path.Join(cfg.RPMBundlingPath(), "script"), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", path.Join(cfg.RPMBundlingPath(), "script"))
	}

	logger.Info().Msg("creating rpm docker template")
	err = templ.Merge(
		"build-rpm-docker-image",
		buildRPMDockerImageTemplate,
		filepath.Join(cfg.RPMBundlingPath(), "Dockerfile"),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "creating rpm/Dockerfile")
	}

	logger.Info().Msg("creating rpm init script")
	err = templ.Merge(
		"build-rpm-init-script",
		buildPythonInitScript,
		filepath.Join(cfg.RPMBundlingPath(), "init.py"),
		struct {
			ServiceName string
		}{
			cfg.ServiceName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating rpm/init.py")
	}
	if err = os.Chmod(path.Join(cfg.RPMBundlingPath(), "init.py"), 0777); err != nil {
		return errors.Wrapf(err, "os.Chmod(%s)", path.Join(cfg.RPMBundlingPath(), "init.py"))
	}

	return nil
}
