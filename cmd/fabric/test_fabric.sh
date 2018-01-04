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

# here we want the GOPATH that points to the (not generated) source code
TEST_CERTS_DIR="$GOPATH/src/github.com/deciphernow/gm-fabric-go/cmd/fabric/test_certs"

TOPDIR="$HOME/fabric_test_dir"

rm -rf $TOPDIR
mkdir $TOPDIR

# here we create a GOPATH that will point to the generated code
GOPATH="$TOPDIR"
mkdir "$GOPATH/src"

TESTDIR="$GOPATH/src/testdir"
mkdir $TESTDIR

SERVICE_NAME="test_service"

# initialize the service
fabric --log-level="debug" --dir="$TESTDIR" --init $SERVICE_NAME

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

# compile the stubs to verify that they are valid
"${SERVICE_NAME}/build_${SERVICE_NAME}_server.sh"
"${SERVICE_NAME}/build_${SERVICE_NAME}_grpc_client.sh"
"${SERVICE_NAME}/build_${SERVICE_NAME}_http_client.sh"
popd

# stuff a client that exercises the methods
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
pushd $TESTDIR
${SERVICE_NAME}/"build_${SERVICE_NAME}_grpc_client.sh"
popd

# stuff a server method that handles a unitary method
cat << METHOD1 > "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_proxy.go"
package methods

import (
    "time"

	"golang.org/x/net/context"

	"github.com/pkg/errors"

    gometrics "github.com/armon/go-metrics"

	pb "testdir/test_service/protobuf"
)

// HelloProxy says "hello" in a form that is handled by the gateway proxy
func (s *serverData) HelloProxy(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {

    defer gometrics.MeasureSince(
		[]string{
			"test_service", // service name
			"HelloProxy",
            "elapsed",
		},
		time.Now(),
	)

	if req.HelloText == "ping" {
        gometrics.IncrCounter(
    		[]string{
    			"test_service", // service name
    			"valid-ping",
    		},
    		1,
    	)
		return &pb.HelloResponse{Text: "pong"}, nil
	}

    gometrics.IncrCounter(
        []string{
            "test_service", // service name
            "invalid-ping",
        },
        1,
    )
	return nil, errors.New("invalid request")
}
METHOD1
gofmt -w "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_proxy.go"

# stuff a server method that handles a stream method
cat << METHOD2 > "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_stream.go"
package methods

import (
    "time"

	gometrics "github.com/armon/go-metrics"

	pb "testdir/test_service/protobuf"
)

// HelloStream says "hello" repeatedly in a stream
func (s *serverData) HelloStream(req *pb.HelloStreamRequest, stream pb.TestService_HelloStreamServer) error {

    defer gometrics.MeasureSince(
		[]string{
			"test_service", // service name
			"HelloStream",
            "elapsed",
		},
		time.Now(),
	)

    for i := int32(0); i < req.Count; i++ {
		stream.Send(&pb.HelloResponse{Text: "pong"})
	}

    return nil
}
METHOD2
gofmt -w "$TESTDIR/$SERVICE_NAME/cmd/server/methods/hello_stream.go"

# compile the server to include the changed methods
# we assume we are running in the test directory
pushd $TESTDIR
"$SERVICE_NAME/build_${SERVICE_NAME}_server.sh"
popd

# stuff our own settings file over the generated one
cat << SETTINGS > "$TESTDIR/$SERVICE_NAME/settings.toml"
# grpc-server
    grpc_server_host = ""
    grpc_server_port = 10000

# metrics-server
    metrics_server_host =  ""
    metrics_server_port = 10001
    metrics_cache_size =  1024
    metrics_uri_path = "/metrics"

# gateway-proxy
    use_gateway_proxy = "true"
    gateway_proxy_host = ""
    gateway_proxy_port = 8080

# tls
    use_tls = true
    ca_cert_path = "$TEST_CERTS_DIR/root.crt"
    server_cert_path = "$TEST_CERTS_DIR/server.localdomain.chain.crt"
    server_key_path = "$TEST_CERTS_DIR/server.localdomain.nopass.key"
	server_cert_name = "server.localdomain"

# statsd
    report_statsd = false
    statsd_host = "127.0.0.1"
    statsd_port = 8125
    statsd_mem_interval = ""

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
$GRPC_CLIENT_BINARY --address=":10000" --test-cert-dir="${TEST_CERTS_DIR}" > "${TOPDIR}/grpc_client.log" 2>&1

HTTP_CLIENT_BINARY="$GOPATH/bin/${SERVICE_NAME}_http_client"

# hit the proxy
$HTTP_CLIENT_BINARY --uri="https://127.0.0.1:8080/acme/services/hello?hello_text=ping" --test-cert-dir="${TEST_CERTS_DIR}" 

# hit the metrics server
$HTTP_CLIENT_BINARY --uri="https://127.0.0.1:10001/metrics" --test-cert-dir="${TEST_CERTS_DIR}"

# stop the server gracefuly
kill $SERVICE_PID

wait
