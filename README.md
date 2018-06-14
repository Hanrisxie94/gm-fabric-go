# Grey Matter - Fabric - Golang

[![CircleCI](https://circleci.com/gh/DecipherNow/gm-fabric-go.svg?style=shield&circle-token=6cfc1eb2506e2e4318762eb78596beddb6efadb4)](https://circleci.com/gh/DecipherNow/gm-fabric-go)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/deciphernow/gm-fabric-go)](https://goreportcard.com/report/github.com/deciphernow/gm-fabric-go)

Grey Matter is an enterprise framework and ecosystem covering common use cases and patterns, so your team can focus on what they do best. This repository contains the source code for the Golang software development kit (SDK) and command line utilities for building fabric enabled microservices.

1. [Prerequisites](#prerequisites)
1. [Installation](#installation)
1. [Creating a Microservice](#creating-a-microservice)
1. [Configuring and Extending a Microservice](#configuring-and-extending-a-microservice)
1. [Warnings](#warnings)
1. [Contributing](#contributing)

## Prerequisites

The following prerequisites must be installed before building any microservce using this SDK. If you are not following the installation instructions herein, you will need to ensure that they are installed prior to developing any microservices.

1.  [Go 1.7+](https://golang.org) (Latest recommended)
1.  [Dep](https://github.com/golang/dep)
1.  [Protobuf Compiler](https://github.com/google/protobuf/releases)

## Installation

While it is possible to develop a fabric enabled microservice from scratch using the packages provided by this repository, it is recommended that you utilize the command line utilities provided to initialize and update your microservices. The following sections provide instructions for installing these utilities on the following systems.

1. [MacOS](#macos)
1. [Other](#other)

### MacOS

For convienence, we provide a Homebrew tap for installing the command line utilities and their prerequisites. If you are not using Homebrew the [other](#other) installation instructions should work for you but there may be discrepancies.

1. If not installed, install Homebrew following the directions at [https://brew.sh](https://brew.sh).
1. Tap the Decipher Homebrew tap to access the formulas.
    ```bash
    brew tap deciphernow/homebrew-decipher
    ```
1. Install Go, Dep, Protobuf and the Fabric command line utilities.  Note that you should only do one of the following commands.
    * To install the latest stable release run the following command:
        ```bash
        brew install deciphernow/homebrew-decipher/fabric
        ```
    * To install the latest development release run the following command:
        ```bash
        brew install --HEAD deciphernow/homebrew-decipher/fabric
        ```
1. Verify that the installation worked.
    ```bash
    fabric --version
    ```
1. Define your `GOPATH`, the path in which Go will download, build and install software, as described [here](https://github.com/golang/go/wiki/SettingGOPATH). For example:
    ```bash
    export GOPATH="${HOME}/go" && echo "export GOPATH=${GOPATH}" >> ~/.bash_profile
    ```
1. Add `${GOPATH}/bin` to `PATH` if you wish to have code you compile available from your shell. This is optional as the compiled code is executable regardless, but you will need to use the full path. For example:
    ```bash
    export PATH="${GOPATH}/bin:${PATH}" && echo "export PATH=\${GOPATH}/bin:\${PATH}" >> ~/.bash_profile
    ```

### Other

The following instructions leverage external sources heavily so please open an issue if you notice link rot or any other discrepancies.

1. Download the Go archive for your platform from [https://golang.org/dl/](https://golang.org/dl/).
1. Follow the installation instructions for your platform from [https://golang.org/doc/install](https://golang.org/dl/).
1. Define your `GOPATH`, the path in which Go will download, build and install software, as described [here](https://github.com/golang/go/wiki/SettingGOPATH). For example:
    ```bash
    export GOPATH="~/go" && echo "export GOPATH=${GOPATH}" >> ~/.bash_profile
    ```
1. Add `${GOPATH}/bin` to `PATH` to have code you compile or install with `go install` available from your shell. For example:
    ```bash
    export PATH="~/go/bin:${PATH}" && echo "export PATH=\${GOPATH}/bin:\${PATH}" >> ~/.bash_profile
    ```
1. Download the `protoc` binary for your platform from [here](https://github.com/google/protobuf/releases)
1. Place the binary somewhere safe in your file system (e.g., `~/bin`).
1. If the directory into which you installed the binary is not already in your `PATH` add it now.
    ```bash
    export PATH="~/bin:~/go/bin:${PATH}" && echo "export PATH=\${GOPATH}/bin:\${PATH}" >> ~/.bash_profile
    ```
1. Install `dep` for dependency resolution.
    ```
    go get -u github.com/golang/dep/cmd/dep
    ```
1. Clone this repository to `${GOPATH}/src/github.com/deciphernow/gm-fabric-go`.
    ```bash
    mkdir -p ${GOPATH}/src/github.com/deciphernow && cd ${GOPATH}/src/github.com/deciphernow && git clone git@github.com:DecipherNow/gm-fabric-go.git
    ```
1. Build and install the command line utilities.
    ```bash
    cd cmd/fabric && dep ensure -v && ./build_fabric.sh
    ```

## Creating a Microservice

### Overview

Once you have installed the utilities you may rapidly initialize and update gRPC based microservices that utilize the Grey Matter Fabric SDK. If you wish to see the full capabilities of the fabric utilities simply run `fabric --help`.

### Initializing a Service

The following instructions assume that we will be creating a service named `exemplar` and it will be hosted at `https://github.com/examples/exemplar.git`.

1. Initialize the service with the `--init` flag.
    ```bash
    fabric --init "exemplar" --template git@github.com:deciphernow/gm-fabric-templates.git//default --dir "${GOPATH}/src/github.com/examples"
    ```
1. The above command will create a new service from the template. Confirm that the service directory exists and is not empty.
    ```bash
    ls -ltra "${GOPATH}/src/github.com/examples/exemplar"
    ```

### Adding Service Methods

As stated previously, Grey Matter Fabric is based upon [gRPC](https://grpc.io/) which uses [Protocol Buffers](https://developers.google.com/protocol-buffers/) interface description language to define the interfaces presented by remote services. To add methods to our exemplar service we need to define those methods in the `${GOPATH}/src/github.com/examples/exemplar/protobuf/exemplar.proto` file.

1. Switch to the parent directory of the service directory that was created by the `--init` command.
    ```bash
    cd ${GOPATH}/src/github.com/examples"
    ```
1. Edit the `${GOPATH}/src/github.com/examples/exemplar/protobuf/exemplar.proto` file to contain the following:
    ```protobuf
    syntax = "proto3";

    package protobuf;

    import "google/api/annotations.proto";

    // Defines the interface implemented and methods exposed by the exemplar service.
    service Exemplar {

        // HelloProxy says 'hello' in a form that is handled by the gateway proxy.
        rpc HelloProxy(HelloRequest) returns (HelloResponse) {

            // Defines the optional REST method and route to be exposed cooresponding to the `HelloProxy` method.
            option (google.api.http) = {
                get: "/examples/services/hello"
            };
        }
    }

    // Defines the request type for the `HelloProxy` method.
    message HelloRequest {
        string hello_text = 1;
    }

    // Defines the response type for the `HelloProxy` method.
    message HelloResponse {
        string text = 1;
    }
    ```
1. Generate the method stubs for the updated definitions.
    ```bash
    fabric --generate "exemplar"
    ```
1. Note that this will add the following files to your service directory.
    ```bash
    ${GOPATH}/go/src/github.com/examples
    └── exemplar
        ├── cmd
        │   └── server
        │       └── methods
        │           ├── hello_proxy.go
        ├── protobuf
        │   ├── exemplar.pb.go
        │   ├── exemplar.pb.gw.go
    ```
1. Edit the `${GOPATH}/src/github.com/examples/exemplar/cmd/server/methods/hello_proxy.go` file with the implementation of the method as follows:
    ```go
    package methods

    import (
        "golang.org/x/net/context"
        "github.com/pkg/errors"
        pb "github.com/examples/exemplar/protobuf"
    )

    // HelloProxy says "hello" in a form that is handled by the gateway proxy
    func (s *serverData) HelloProxy(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
        if req.HelloText == "ping" {
            return &pb.HelloResponse{Text: "pong"}, nil
        }
        return nil, errors.New("invalid request")
    }
    ```

Note that any time you make changes to your proto files (e.g., `${GOPATH}/src/github.com/examples/exemplar/protobuf/exemplar.proto`) you should rerun the `fabric --generate` command (e.g., `fabric --dir="${GOPATH}/src/github.com/examples" --generate "exemplar"`) to ensure that the generated code is up to date with your proto files. Any changes you have made to the method implementations will not be overwritten.

### Building and Running the Service

Once you are happy with your implementation of your service you may build and run the service.

1. You must run the build scripts from the current directory.
    ```bash
    cd ${GOPATH}/src/github.com/examples/exemplar
    ```
1. Build the service.
    ```bash
    ./build_server.sh
    ```
1. Run the service.
    ```bash
    ${GOPATH}/bin/exemplar --config ${GOPATH}/src/github.com/examples/exemplar/settings.toml
    ```
1. Make a simple REST call to the service.
   ```bash
   curl http://localhost:8080/examples/services/hello?hello_text=ping
   ```
1. Additionally, you may retrive your service's metrics.
   ```bash
   curl http://localhost:10001/metrics
   ```

### Testing the Service

The generated service also provides stubs for creating tests of your service using the gRPC client that was generated. The following instructions will provide an example of adding tests to the exemplar service.

1. Edit the test stub located here `${GOPATH}/src/github.com/examples/exemplar/cmd/grpc_client/test_grpc.go` as follows:
    ```go
    package main

    import (
        "golang.org/x/net/context"

        "github.com/pkg/errors"
        "github.com/rs/zerolog"

        pb "github.com/examples/exemplar/protobuf"
    )

    func runTest(logger zerolog.Logger, client pb.ExemplarClient) error {
        req := pb.HelloRequest{HelloText: "ping"}

        resp, err := client.HelloProxy(context.Background(), &req)
        if err != nil {
            return errors.Wrap(err, "HelloRequest")
        }
        logger.Info().Str("response", resp.Text).Msg("")

        return nil
    }
    ```
1. You must run the build scripts from the current directory.
    ```bash
    cd ${GOPATH}/src/github.com/examples/exemplar
    ```
1. Build the tests.
    ```bash
     ./build_exemplar_grpc_client.sh
     ./build_exemplar_http_client.sh
    ```
1. Run the exemplar service if not running.
   ```bash
    ${GOPATH}/bin/exemplar --config ${GOPATH}/src/github.com/examples/exemplar/settings.toml
    ```
1. Run the grpc tests.
   ```bash
   ${GOPATH}/bin/exemplar_grpc_client --address localhost:10000
   ```

## Configuring and Extending a Microservice

### Overview

With a basic microservice implemented you may extend the capabilities of that service or explore more advanced configuration options. Instructions for this are provided in the packages for those capabilities.

### Packages

| Package                            | Description                                                        |
|------------------------------------|--------------------------------------------------------------------|
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

### OpenSSL on OS X or macOS

Due to Apple changing their security code and removing OpenSSL header files from `/usr/bin/openssl/` please follow the steps bellow to solve any issues involving OpenSSL on OS X 10.11 (El Capitan) or newer.

*Note: if you have openssl already installed to `/usr/local/bin/` you will want to back that up with `cd /usr/local/bin && mv openssl openssl.bak`*

1. Make a backup of `/usr/local/bin/openssl` if it exists.
    ```bash
    if [ -f /usr/local/bin/openssl ];then mv /usr/local/bin/openssl /usr/local/bin/openssl.bak; fi
    ```
1.  Install Homebrew if not already installed.
    ```bash
    which brew || /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    ```
1.  Install OpenSSL and PKG-Config. Homebrew may warn that this is keg-only. This is not an issue, just verify that `/usr/local/Cellar/openssl` exists.
    ```bash
    brew install openssl pkg-config
    ```
1.  Symlink the installed OpenSSL to `/usr/local/bin/openssl`.
    ```bash
    ln -s /usr/local/Cellar/openssl/{openssl_version}/bin/openssl /usr/local/bin/openssl
    ```
1.  Add the new OpenSSL path to the `PKG_CONFIG_PATH`.
    ```bash
    export PKG_CONFIG_PATH="$(brew --prefix openssl)/lib/pkgconfig" && echo "export PKG_CONFIG_PATH=${PKG_CONFIG_PATH}" >> ${HOME}/.bash_profile
    ```
## Contributing

1. Fork it
1. Create your feature branch (`git checkout -b my-new-feature`)
1. Commit your changes (`git commit -am 'Add some feature'`)
1. Push to the branch (`git push origin my-new-feature`)
1. Create new Pull Request
