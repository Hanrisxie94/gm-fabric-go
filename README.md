# GM Fabric Golang SDK
[![CircleCI](https://circleci.com/gh/DecipherNow/gm-fabric-go.svg?style=shield&circle-token=6cfc1eb2506e2e4318762eb78596beddb6efadb4)](https://circleci.com/gh/DecipherNow/gm-fabric-go)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go)

Grey Matter Fabric Golang Software Development Kit

1.  [Getting Started](#getting-started)
2.  [Packages](#packages)
3.  [Warnings](#warnings)

## Prerequisites
1.  [Go 1.7+](https://golang.org) (Latest recommended)
2.  [Dep](https://github.com/golang/dep)
3.  [Protobuf Compiler](https://github.com/google/protobuf/releases)

# Getting Started

## Install and Setup Go

### With Homebrew

```bash
brew install go
```

### With the Installer Package

1.  Download the installer for your platform from [https://golang.org/dl/](https://golang.org/dl/)
2.  Follow the install prompts

When Go installs packages, it puts the artifacts in the following tree:
```bash
$GOPATH   <-- you set this environment variable
├── bin   <-- executables go here
├── pkg   <-- object files (*.a) go here
└── src   <-- source files go here
```

To get started, decide where you want your $GOPATH to be. For example:
```bash
export GOPATH=$HOME/go
```

To make go programs available in your shell:
```bash
PATH=$GOPATH/bin:$PATH
```

*Note*: Both need to be added to `$HOME/.bash_profile` for persistence across shell sessions.

## Install and Setup The Protobuf Compiler

### With Homebrew

```bash
brew install protobuf
```

### Without Homebrew

1. Download the `protoc` binary for your OS: [release page](https://github.com/google/protobuf/releases)
2. Place the binary somewhere safe in your file system.
3. Add the new binary to your path. Ex:
```bash
PATH=$PATH:~/Developer/protoc-3.3.0-osx-x86_64/bin
```
*Note*: Add the new path to `$HOME/.bash_profile` for persistence across shell sessions.

## Build A Micro-Service

### Installing The Grey Matter Service Generator
1. `cd` into `cmd/fabric`
2. Run `dep ensure -v`
3. Run `./build_fabric.sh`

### Overview
A quick outline of the Grey Matter micro-service generator:
```
Usage of fabric:
  --dir string
    	path to the directory containing the service. Default: cwd
  --init <service-name>
        initialize service
        run once to initialize service
  --generate <service-name>
    	generate protobuf methods for service
        run repeatedly while changing the protobuf definition (.proto)
```
If you are interested in taking a peek under the hood, the source code for the service generator is located under `cmd/fabric`.

### Initialize a Skeleton
1. It might be useful to think of a name for your service beforehand...
2. Execute the `init` function of the Grey Matter service generator:
```bash
fabric --init <service_name>
```
A contrived example:
```bash
fabric --dir=$GOPATH/src/github.com/<organization-name> --init "test_service"
```
If all was successful, the generator will spit out the following directory structure:
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
*Note*: `test_service.proto` is a stub. You will need to edit this for further generation.
### Iterative Generation
Once you have initialized your skeleton and edited your protobuf file (outlining your gRPC micro-service), you will need to run the following commands to have a fully working micro-service:
```bash
fabric --generate <service-name>
```
*Note*: if you `cd`'d into the directory of your service skeleton, you will need to back out one directory and run this command since the default dir that gm-servgen uses is `cwd`. Otherwise, you may specify the dir with this flag:
```bash
fabric --dir=$GOPATH/src/src/github.com/<organization-name> --generate <service-name>
```
*Note*: everytime you edit the protobuf file, you will need to regenerate with the command above. Hence why this is an iterative process.

### Example
Edit (or replace) the stub `test_service.proto`
```protobuf
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
### Generate
```bash
fabric --dir=$GOPATH/src/src/github.com/<organization-name> --generate "test_service"
```

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
Each rpc request in `test_service.proto` gets its own method file, such as
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
You can edit the generated method to add functionality. When you rerun `--generate` we will **not** write over your changes.  

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
*Note* above that you can include your own metrics in the server methods, such as:
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
### Build The Test Server
After adding code to methods, you can build the server with ```./build_<service name>_server.sh```
```bash
#! /bin/bash

set -euxo pipefail

pushd /tmp/testtopdir/src/testdir/test_service/cmd/server
go build -o=$GOPATH/bin/test_service
popd
```
## Test client
We also generate the stub of a grpc client for testing.
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
You can add code to test changes to server methods:
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
{
    "text":"pong"
}
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

## Packages

| Package                            | Description                                                        |
|------------------------------------|--------------------------------------------------------------------|
| [cloudwatch](cloudwatch/README.md) | Auto-scale abilities using GM Fabric metrics and amazon cloudwatch |
| [consul](consul/README.md)         | Easy integration with Hashicorps service discovery system          |
| [dbutil](dbutil/README.md)         | Utilities for configuration/interaction with Mongo and Redis       |
| [events](events/README.md)         | Prototyped event streaming                                         |
| [fabric](https://github.com/DecipherNow/gm-fabric-go/cmd/fabric) | GM Fabric golang service generator (gRPC) |
| [gk](gk/README.md)                 | Gatekeeper service announcement utility                            |
| [metrics](metrics/README.md)       | GM Fabric metrics (HTTP, gRPC)                                     |
| [middleware](middleware/README.md) | HTTP middleware helpers                                            |
| [oauth](oauth/README.md)           | OAuth 2.0 authorization code (recommended use with coreos/dex)     |
| [rpcutil](rpcutil/README.md)       | Utility functions for GM Fabric gRPC services                      |
| [sds](sds/README.md)               | Service discovery and announcement (GM Fabric 2.0)                 |
| [tlsutil](tlsutil/README.md)       | TLS utility functions for easy integration with 2-way SSL          |
| [zkutil](zkutil/README.md)         | Helper functions for zookeeper/gatekeeper service announcement     |

## Warnings

Due to Apple changing their security code and removing OpenSSL header files from `/usr/bin/openssl/` please follow the steps bellow to solve any issues involving OpenSSL on OS X 10.11 (El Capitan) or newer.

*Note: if you have openssl already installed to `/usr/local/bin/` you will want to back that up with these commands. If not, proceed to step 1*
```bash
cd /usr/local/bin
mv openssl openssl.orig
```

1.  Install Homebrew

```bash
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```

2.  Install OpenSSL and PKG-Config

```bash
brew install openssl
brew install pkg-config
```
*Note*: Homebrew may say this is keg-only, do not worry. That is just a warning. You can verify if the install was correct by looking in `/usr/local/Cellar/openssl`. If so, Proceed to step 3

3.  Create SymLinks

```bash
ln -s /usr/local/Cellar/openssl/{openssl_version}/bin/openssl /usr/local/bin/openssl
```

*Note*: This will tell Mac OS to use the homebrew version of OpenSSL which includes the header and necessary development files

4.  Update `~/.bash_profile` or `~/.bashrc`

```bash
export PKG_CONFIG_PATH="$(brew --prefix openssl)/lib/pkgconfig"
```
