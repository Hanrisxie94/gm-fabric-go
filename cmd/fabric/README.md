# Service Generator
## Getting Started
 * Have `go` installed and in your path
   * https://golang.org/dl/
 * Have the `dep` binary in your path
    * https://github.com/golang/dep
 * Have the `GOPATH` environment variable set
    * Go 1.8 introduces a default GOPATH: `$HOME/go` on Unix-like systems
    * Please declare the `GOPATH` environment variable anyway. Add to `$HOME/.profile`
    ```bash
    export GOPATH="$HOME/go"
    export PATH=$PATH:$GOPATH/bin
    ```
    * Technically, the `GOPATH` environment variable could contain a list of paths:
    ```bash
    export GOPATH="$HOME/go:/some/random/path"
    ```
    Please don't do that.
 * Have a top level folder under `$GOPATH/src`.   
   * Many developers work under the folder that would be populated by `go get` https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies
   * Example: `$GOPATH/src/github.com/<organization-name>`
 * Have access to GitHub
    * You will like it a lot better if it's public key access
    * https://help.github.com/articles/connecting-to-github-with-ssh/)
 * Have a recent version of `protoc` available.
   * https://gist.github.com/sofyanhadia/37787e5ed098c97919b8c593f0ec44d8#file-install-protobuf-3-on-ubuntu

### Examples
Examples in this document come from the test script ```gm-fabric-go/cmd/fabric/test_fabric.sh```. You may be interested in reading and/or running it.

## Usage
```
Usage of gm-fabric-go/cmd/fabric:
  -dir string
    	path to the directory containing the service. Default: cwd
  -init <service-name>
        initialize service
        run once to initialize service
  -generate <service-name>
    	generate protobuf methods for service
        run repeatedly while changing the protobuf definition (.proto)
```
## Initialize Service
`gm-fabric-go/cmd/fabric --init <service-name>`

for example: `gm-fabric-go/cmd/fabric --dir=$GOPATH/src/github.com/<organization-name> --init "test_service"`

The program will create the following directory structure
```
$HOME/<user-name>/go/src/github.com/<organization-name>
└── test_service
    ├── build_test_service_grpc_client.sh
    ├── build_test_service_server.sh
    ├── cmd
    │   ├── grpc_client
    │   │   ├── main.go
    │   │   └── test_grpc.go
    │   └── server
    │       ├── config
    │       │   └── config.go
    │       ├── main.go
    │       └── methods
    │           └── new_server.go
    ├── Gopkg.lock
    ├── Gopkg.toml
    ├── protobuf
    │   └── test_service.proto
    ├── settings.toml
    └── vendor
```
`test_service.proto` is a stub:
```
syntax = "proto3";

package protobuf;

import "google/api/annotations.proto";

// Interface exported by the server.
service TestService {
}
```
## Generate Service
This is an iterative process:
 * Add method and data definitions to the protocol buffer definition file (.proto)
 * `gm-fabric-go/cmd/fabric --generate <service-name>`
 * Add working code to the generated methods

### Edit Protocol Buffer Definition File
for example: edit (or replace) the stub `test_service.proto`)
```
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
}
message HelloRequest {
    string hello_text = 1;
}

message HelloResponse {
	string text = 1;
}
```
### run --generate
`gm-fabric-go/cmd/fabric --dir=$GOPATH/src/src/github.com/<organization-name> --generate "test_service"`


This produces generated files:
```
$HOME/<user-name>/go/src/github.com/<organization-name>
└── test_service
    ├── cmd
    │   └── server
    │       └── methods
    │           ├── hello_proxy.go
    ├── protobuf
    │   ├── test_service.pb.go
    │   ├── test_service.pb.gw.go
```
#### Method stub
each rpc request in `test_service.proto` gets its own method file, such as
`hello_proxy.go`. The generated files are initially stubs:
```golang
package methods

import (
    "fmt"

    "golang.org/x/net/context"

	pb "testdir/test_service/protobuf"
)

// HelloProxy says "hello" in a form that is handled by the gateway proxy
func (s *serverData) HelloProxy(context.Context, *pb.HelloRequest) (*pb.HelloResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
```
### Edit method stub
You can edit the generated method to add functionality.

When you rerun ```--generate``` we will **not** write over your changes.  

For example:
```go
package methods

import (
    "time"

	"golang.org/x/net/context"

	"github.com/pkg/errors"

    gometrics "github.com/armon/go-metrics"

	pb "testdir/test_service/protobuf"
)

// HelloProxy says &#39;hello&#39; in a form that is handled by the gateway proxy
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
```
Note above that you can include your own metrics in the server methods, such as:
```go
defer gometrics.MeasureSince(
    []string{
        "test_service", // service name
        "HelloProxy",
        "elapsed",
    },
    time.Now(),
)
```
These will be reported by the metrics server
```json
"go_metrics/valid-ping": 2.000000,
"go_metrics/HelloProxy/elapsed": 0.004140
```
### Build Server
After adding code to methods, you can build the server with ```build_<service name>_server.sh```
```bash
#! /bin/bash

set -euxo pipefail

pushd /tmp/testtopdir/src/testdir/test_service/cmd/server
go build -o=$GOPATH/bin/test_service
popd
```
## Test client
we also generate the stub of a grpc client for testing.
```
.
└── test_service
    ├── cmd
    │   ├── grpc_client
    │   │   ├── main.go
    │   │   └── test_grpc.go

```
```go
package main

import (
	"golang.org/x/net/context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	pb "testdir/test_service/protobuf"
)

func runTest(logger zerolog.Logger, client pb.TestServiceClient) error {
    return errors.New("not implemented")
}
```
You can add code to test changes to server methods
```go
package main

import (
	"golang.org/x/net/context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	pb "testdir/test_service/protobuf"
)

func runTest(logger zerolog.Logger, client pb.TestServiceClient) error {
	req := pb.HelloRequest{HelloText: "ping"}
	resp, err := client.HelloProxy(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "HelloRequest")
	}
	logger.Info().Str("response", resp.Text).Msg("")
	return nil
}
```
### Build Client
After adding test code, you can build the client with ```build_<service name>_grpc_client.sh```
```bash
#! /bin/bash

set -euxo pipefail

pushd /tmp/testtopdir/src/testdir/test_service/cmd/grpc_client
go build -o=$GOPATH/bin/test_service_grpc_client
popd
```
### Test Gateway
You don't need a client to test the gateway proxy, you can do it with ```curl```
```bash
curl 'http://127.0.0.1:8080/acme/services/hello?hello_text=ping'
```
```json
{"text":"pong"}
```
### Test Metrics Server
You can also use ```curl``` to query the metrics server.  The result is a JSON array:
```bash
curl http://127.0.0.1:10001/metrics
```
```json
{
	"Total/requests": 2,
	"HTTP/requests": 0,
	"HTTPS/requests": 0,
	"RPC/requests": 2,
	"RPC_TLS/requests": 0,
	"function/HelloProxy/requests": 2,


    "go_metrics/runtime/alloc_bytes": 1394328.000000,
	"go_metrics/valid-ping": 2.000000,
	"go_metrics/HelloProxy/elapsed": 0.004140
}
```
