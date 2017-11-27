# RPCUtil
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/rpcutil)

Utility functions for use with the gRPC runtime

1. [Prerequisites](#prerequisites)
2. [Install](#install)
3. [Usage](#usage)

## Prerequisites

1. [Go](https://golang.org/) (1.8+ recommended)
2. [Protobuf](https://github.com/golang/protobuf)

## Install

1. Add the dependency with [golang/dep](https://github,com/golang/dep)
```bash
dep ensure -v -add github.com/deciphernow/gm-fabric-go/rpcutil
```

## Usage

1. Import the the package into your file:
```go
import (
    "github.com/deciphernow/gm-fabric-go/rpcutil"
)
```

2. Use whichever function fits your use case. In this example we are using a function that passes all HTTP headers from a gRPC gateway directly without any string manipulation.
```go
grpcRuntime.WithIncomingHeaderMatcher(rpcutil.MatchHTTPHeaders)
```

*Note*

These utility functions are written to be passed around according to the functional opts pattern. The Go gRPC development kit follows this pattern when injecting configuration options.
