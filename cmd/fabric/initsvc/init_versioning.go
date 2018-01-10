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
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
)

func initVersioning(
	cfg config.Config,
	logger zerolog.Logger,
) error {
	var cwd string
	var cmd *exec.Cmd
	var err error

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

	cmd = exec.Command("dep", "init", "-v")
	if err = executeCommand(cmd, logger); err != nil {
		return errors.Wrap(err, "dep init")
	}

	// now, due to the vagaries of the grpc gateway plugin, we must hack
	// Gopkg.toml and regenerate some vendor stuff
	if err = hackGopkgToml(cfg, logger); err != nil {
		return errors.Wrap(err, "hackGopkgToml")
	}

	// feed our hacked Gopkg.toml to dep
	cmd = exec.Command("dep", "ensure", "-v")
	if err = executeCommand(cmd, logger); err != nil {
		return errors.Wrapf(err, "dep ensure")
	}

	// do an "install" of required binary tools FROM INSIDE VENDOR;
	if err = installProtocGen(cfg, logger); err != nil {
		return errors.Wrap(err, "installProtocGen")
	}

	// do an "install" of required binary tools FROM INSIDE VENDOR;
	if err = installGateway(cfg, logger); err != nil {
		return errors.Wrap(err, "installGateway")
	}

	logger.Debug().Msg("end initVersioning")
	return nil
}

// hackGopkgToml appends some text to Gopkg.toml to readjust some dependencies
func hackGopkgToml(cfg config.Config, logger zerolog.Logger) error {
	const fixText = `

	# appended by gm-fabric-go/cmd/fabric
	[[override]]
	  name = "github.com/grpc-ecosystem/grpc-gateway"
	  branch = "master"
	`
	var inDdata []byte
	var outData []byte
	var err error

	gpkgPath := filepath.Join(cfg.ServicePath(), "Gopkg.toml")

	if inDdata, err = ioutil.ReadFile(gpkgPath); err != nil {
		return errors.Wrapf(err, "ioutil.ReadFile(%s)", gpkgPath)
	}

	outData = bytes.Join([][]byte{inDdata, []byte(fixText)}, nil)

	if err = ioutil.WriteFile(gpkgPath, outData, 0777); err != nil {
		return errors.Wrapf(err, "ioutil.WriteFile(%s", gpkgPath)
	}

	return nil
}

func installProtocGen(cfg config.Config, logger zerolog.Logger) error {
	var cwd string
	var cmd *exec.Cmd
	var err error

	pDir := path.Join(
		cfg.VendorPath(),
		"github.com",
		"golang",
		"protobuf",
		cfg.ProtocGenGoPluginName(),
	)

	logger.Debug().Str("install", pDir).Msg("")

	if cwd, err = os.Getwd(); err != nil {
		return errors.Wrap(err, "os.Getwd()")
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			panic(err)
		}
	}()
	if err = os.Chdir(pDir); err != nil {
		return errors.Wrapf(err, "os.Chdir(%s)", pDir)
	}

	cmd = exec.Command("go", "install", "-v")
	if err = executeCommand(cmd, logger); err != nil {
		return errors.Wrapf(err, "go install gateway")
	}

	return nil
}

func installGateway(cfg config.Config, logger zerolog.Logger) error {
	var cwd string
	var cmd *exec.Cmd
	var op []byte
	var err error

	gatewayDir := path.Join(
		cfg.VendorPath(),
		"github.com",
		"grpc-ecosystem",
		"grpc-gateway",
		cfg.ProtocGenGatewayPluginName(),
	)

	logger.Debug().Str("install", gatewayDir).Msg("")

	if cwd, err = os.Getwd(); err != nil {
		return errors.Wrap(err, "os.Getwd()")
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			panic(err)
		}
	}()
	if err = os.Chdir(gatewayDir); err != nil {
		return errors.Wrapf(err, "os.Chdir(%s)", gatewayDir)
	}

	cmd = exec.Command("go", "install", "-v")
	if op, err = cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "go install gateway; %s", string(op))
	}

	return nil
}

// helper execute function to capture output to os.Stdout so commands don't look like they're hanging
func executeCommand(cmd *exec.Cmd, logger zerolog.Logger) error {
	op, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// create a scanner to read the output stream of the command
	scanner := bufio.NewScanner(op)
	go func() {
		// loop over the scanner to read the output and print
		for scanner.Scan() {
			logger.Info().Msgf("%s", scanner.Text())
		}
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}
