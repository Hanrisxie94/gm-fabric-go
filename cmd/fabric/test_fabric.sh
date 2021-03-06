#!/bin/bash

# Copyright 2017 Decipher Technology Studios LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euxo pipefail

CERTS_PATH="https://deciphernow.github.io/gm-fabric-documentation/certificates"
TOPDIR="$HOME/fabric_test_dir"

rm -rf $TOPDIR
mkdir $TOPDIR

# here we create a GOPATH that will point to the generated code
GOPATH="$TOPDIR"
mkdir "$GOPATH/src"

TESTDIR="$GOPATH/src/testdir"
mkdir $TESTDIR

SERVICE_NAME="test_service"

TEMPLATES="${1:-}"

# initialize the service
fabric --log-level="debug" $TEMPLATES --dir="$TESTDIR" --init $SERVICE_NAME

# add method to the protocol buf definition by stuffing a whole new
# file from a 'here' document
cat << PROTO1 > "$TESTDIR/$SERVICE_NAME/protobuf/test_service.proto"
syntax = "proto3";

package protobuf;

import "google/api/annotations.proto";

// Interface exported by the server.
service TestService {
    // HelloProxy says 'hello' in a form that is handled by the gateway proxy
	rpc HelloProxy(HelloRequest) returns (HelloResponse) {
		option (google.api.http) = {
            get: "/acme/services/hello"
        };
	}

    // HelloStream returns multiple replies
	rpc HelloStream(HelloStreamRequest) returns (stream HelloResponse) {}

}
message HelloRequest {
    string hello_text = 1;
}

message HelloStreamRequest {
    string hello_text = 1;
    int32 count = 2;
}

message HelloResponse {
	string text = 1;
}

PROTO1

# run again to generate the protobuf files and our method stub(s)
# we assume we are running in the test directory
pushd $TESTDIR
fabric --log-level="debug" --generate $SERVICE_NAME
popd


pushd "$TESTDIR/${SERVICE_NAME}"
# compile the stubs to verify that they are valid
"./build_server.sh"
"./build_grpc_client.sh"
"./build_http_client.sh"
popd

# stuff a client that exercises the methodsgit@github.com:deciphernow/gm-fabric-templates.git//default
cat << CLIENT1 > "$TESTDIR/$SERVICE_NAME/cmd/grpc_client/test_grpc.go"
package main

import (
    "io"

	"golang.org/x/net/context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	pb "testdir/test_service/protobuf"
)

func runTest(logger zerolog.Logger, client pb.TestServiceClient) error {
	var err error

	if err = testUnitaryMethod(logger, client); err != nil {
		return errors.Wrap(err, "testUnitaryMethod")
	}

	if err = testStreamMethod(logger, client); err != nil {
		return errors.Wrap(err, "testStreamMethod")
	}

	return nil
}

func testUnitaryMethod(logger zerolog.Logger, client pb.TestServiceClient) error {
	req := pb.HelloRequest{HelloText: "ping"}
	resp, err := client.HelloProxy(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "HelloRequest")
	}
	logger.Info().Str("response", resp.Text).Msg("")

	return nil
}

func testStreamMethod(logger zerolog.Logger, client pb.TestServiceClient) error {
	var hsc pb.TestService_HelloStreamClient
    var count int
	var err error

	req := pb.HelloStreamRequest{HelloText: "ping", Count: 5}

	hsc, err = client.HelloStream(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "client.HelloStream")
	}

	for loop := true; loop; {
		var resp *pb.HelloResponse

		if resp, err = hsc.Recv(); err != nil {
			if err == io.EOF {
				loop = false
			} else {
				return errors.Wrap(err, "hsc.Recv()")
			}
		} else {
			count++
			logger.Info().Int("count", count).Str("response", resp.Text).Msg("")
		}
	}

    return nil
}
CLIENT1
gofmt -w "$TESTDIR/$SERVICE_NAME/cmd/grpc_client/test_grpc.go"

# compile the gRPC client again, this  time with real code
# we assume we are running in the test directory
pushd "${TESTDIR}/${SERVICE_NAME}"
"./build_grpc_client.sh"
popd

# stuff a server method that handles a unitary method
cat << METHOD1 > "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_proxy.go"
package methods

import (
	"golang.org/x/net/context"

	"github.com/pkg/errors"

	pb "testdir/test_service/protobuf"
)

// HelloProxy says "hello" in a form that is handled by the gateway proxy
func (s *serverData) HelloProxy(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {

	if req.HelloText == "ping" {
		return &pb.HelloResponse{Text: "pong"}, nil
	}

	return nil, errors.New("invalid request")
}
METHOD1
gofmt -w "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_proxy.go"

# stuff a server method that handles a stream method
cat << METHOD2 > "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_stream.go"
package methods

import (
	pb "testdir/test_service/protobuf"
)

// HelloStream says "hello" repeatedly in a stream
func (s *serverData) HelloStream(req *pb.HelloStreamRequest, stream pb.TestService_HelloStreamServer) error {

    for i := int32(0); i < req.Count; i++ {
		stream.Send(&pb.HelloResponse{Text: "pong"})
	}

    return nil
}
METHOD2
gofmt -w "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_stream.go"

# compile the server to include the changed methods
# we assume we are running in the test directory
# get the 
pushd "${TESTDIR}/${SERVICE_NAME}"
"./build_server.sh"
popd

# get our test certificates
curl "$CERTS_PATH/localhost.crt" > "$TESTDIR/$SERVICE_NAME/localhost.crt"
curl "$CERTS_PATH/localhost.key" > "$TESTDIR/$SERVICE_NAME/localhost.key"
curl "$CERTS_PATH/intermediate.crt" > "$TESTDIR/$SERVICE_NAME/intermediate.crt"

# stuff our own settings file over the generated one
cat << SETTINGS > "$TESTDIR/$SERVICE_NAME/settings.toml"
# grpc-server
    grpc_use_tls = true
    grpc_server_host = ""
    grpc_server_port = 10000

# gateway-proxy
    gateway_use_tls = true
    use_gateway_proxy = true
    gateway_proxy_host = ""
    gateway_proxy_port = 8080
    gateway_serve_anonymous_arrays = true

# tls
	ca_cert_path = "$TESTDIR/$SERVICE_NAME/intermediate.crt"
	server_cert_path = "$TESTDIR/$SERVICE_NAME/localhost.crt"
	server_key_path = "$TESTDIR/$SERVICE_NAME/localhost.key"
	server_cert_name = "localhost"

# misc
    verbose_logging = true

SETTINGS

# run the server in background
SERVICE_BINARY="$GOPATH/bin/$SERVICE_NAME"
SERVICE_CONFIG="$TESTDIR/$SERVICE_NAME/settings.toml"

$SERVICE_BINARY --config="$SERVICE_CONFIG" > "${TOPDIR}/server.log" 2>&1 &
SERVICE_PID=$!

ps -p $SERVICE_PID

# run the grpc client as a test
GRPC_CLIENT_BINARY="$GOPATH/bin/${SERVICE_NAME}_grpc_client"
$GRPC_CLIENT_BINARY --address=":10000" --test-cert-dir="$TESTDIR/$SERVICE_NAME" > "${TOPDIR}/grpc_client.log" 2>&1

HTTP_CLIENT_BINARY="$GOPATH/bin/${SERVICE_NAME}_http_client"

# hit the proxy
$HTTP_CLIENT_BINARY --uri="https://127.0.0.1:8080/acme/services/hello?hello_text=ping" --test-cert-dir="$TESTDIR/$SERVICE_NAME" 

# stop the server gracefuly

kill $SERVICE_PID

wait
