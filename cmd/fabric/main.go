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

package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/genproto"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/initsvc"
)

func main() {
	os.Exit(run())
}

func run() int {

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		logger.Error().Err(err).Msg("config.Load()")
		return 1
	}

	zerolog.SetGlobalLevel(cfg.LogLevel)

	logger.Debug().Str("GOPATH", os.Getenv("GOPATH")).Msg("")
	logger.Debug().Str("GOBIN", os.Getenv("GOBIN")).Msg("")
	logger.Debug().Str("version", cfg.Version).Msg("")

	switch cfg.Op {
	case config.ShowVersion:
		fmt.Println(cfg.Version)
	case config.Init:
		if err = initsvc.InitService(cfg, logger); err != nil {
			logger.Error().AnErr("initService", err).Msg("")
			return 1
		}
	case config.Generate:
		if err = genproto.GenerateProtobuf(cfg, logger); err != nil {
			logger.Error().AnErr("generateProtobuf", err).Msg("")
			return 1
		}
	default:
		logger.Error().Int("invalid config.Operation", int(cfg.Op)).Msg("")
		return 1
	}

	return 0
}
