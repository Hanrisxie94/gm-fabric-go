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

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/cmd/gm-servgen/config"
	"github.com/deciphernow/gm-fabric-go/cmd/gm-servgen/templ"
)

func createGRPCClient(
	cfg config.Config,
	logger zerolog.Logger,
) error {
	var err error

	logger.Info().Msg("creating client main.go")
	err = templ.Merge(
		"client",
		grpcClientTemplate,
		filepath.Join(cfg.GRPCClientPath(), "main.go"),
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
		return errors.Wrap(err, "creating client main.go")
	}

	logger.Info().Msg("creating test_grpc.go")
	err = templ.Merge(
		"test_grpc",
		grpcClientTestTemplate,
		filepath.Join(cfg.GRPCClientPath(), "test_grpc.go"),
		struct {
			GoServiceName string
			PBImport      string
		}{
			cfg.GoServiceName(),
			cfg.PBImportPath(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "creating test_grpc.go")
	}

	return nil
}
