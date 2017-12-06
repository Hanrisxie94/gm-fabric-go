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

var grpcClientTemplate = `package main

import (
    "os"

    "github.com/pkg/errors"
	"github.com/rs/zerolog"
    "github.com/spf13/pflag"

	"google.golang.org/grpc"

    pb "{{.PBImport}}"
)

func main() {
    os.Exit(run())
}

func run() int {
    var grpcServerAddress string
    var client pb.{{.GoServiceName}}Client
    var err error

    logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	pflag.StringVar(
		&grpcServerAddress,
		"address",
		"",
		"address of grpc server",
	)
	pflag.Parse()
    if grpcServerAddress == "" {
        logger.Error().Msg("You must specify server address. (--address)")
        return 1
    }

    logger.Info().Str("grpc-client", "{{.ServiceName}}").
        Str("address", grpcServerAddress).Msg("starting")

    if client, err = newClient(grpcServerAddress); err != nil {
        logger.Error().AnErr("newClient", err).Msg("")
        return 1
	}

    if err = runTest(logger, client); err != nil {
        logger.Error().AnErr("runTest", err).Msg("")
        return 1
	}

    logger.Info().Str("grpc client", "{{.ServiceName}}").Msg("terminating normally")
    return 0
}

func newClient(serverAddress string) (pb.{{.GoServiceName}}Client, error) {
	var opts []grpc.DialOption
	var conn *grpc.ClientConn
	var err error

	opts = append(opts, grpc.WithInsecure())
	conn, err = grpc.Dial(serverAddress, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc.Dial(%s", serverAddress)
	}

	return pb.New{{.GoServiceName}}Client(conn), nil
}
`
