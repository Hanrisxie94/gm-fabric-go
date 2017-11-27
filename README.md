# GM Fabric Golang SDK
[![CircleCI](https://circleci.com/gh/DecipherNow/gm-fabric-go.svg?style=shield&circle-token=ac1acca0b338abb9fa0f67736e6e26e1832321db)](https://circleci.com/gh/DecipherNow/gm-fabric-go)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go)

Grey Matter Fabric Golang Software Development Kit

1.  [Packages](#packages)
2.  [Install](#install)
3.  [Warnings](#warnings)

## Prerequisites
1.  [Go 1.7+](https://golang.org) (Latest recommended)

## Packages

| Package                            | Description                                                    |
|------------------------------------|----------------------------------------------------------------|
| [consul](consul/README.md)         | Easy integration with Hashicorps service discovery system      |
| [dbutil](dbutil/README.md)         | Utilities for configuration/interaction with Mongo and Redis   |
| [events](events/README.md)         | Prototyped event streaming                                     |
| [gk](gk/README.md)                 | Gatekeeper service announcement utility                        |
| [metrics](metrics/README.md)       | GM Fabric metrics (HTTP, gRPC)                                 |
| [middleware](middleware/README.md) | HTTP middleware helpers                                        |
| [oauth](oauth/README.md)           | OAuth 2.0 authorization code (recommended use with coreos/dex) |
| [rpcutil](rpcutil/README.md)       | Utility functions for GM Fabric gRPC services                  |
| [sds](sds/README.md)               | Service discovery and announcement (GM Fabric 2.0)             |
| [srvgen](srvgen/README.md)         | GM Fabric golang service generator (gRPC)                      |
| [tlsutil](tlsutil/README.md)     | TLS utility functions for easy integration with 2-way SSL      |
| [zkutil](zkutil/README.md)         | Helper functions for zookeeper/gatekeeper service announcement |
| [cloudwatch](cloudwatch/README.md) | Auto-scale abilities using GM Fabric metrics and amazon cloudwatch |

## Install

We recommend that you use [golang/dep](https://github.com/golang/dep) for dependency management.

### Alternative

1.  Clone the repo into `$GOPATH/src/github.com/deciphernow`
2.  Import the appropriate package into your code:
```go
import (
    "github.com/deciphernow/gm-fabric-go/metrics"
)
```

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

2.  Install OpenSSL and Pkg-Config

    ```bash
    brew install openssl
    brew install pkg-config
    ```
    *Note: Homebrew may say this is keg-only, do not worry. That is just a warning. You can verify if the install was correct by looking in `/usr/local/Cellar/openssl`. If so, Proceed to step 3*
3.  Create SymLinks

    ```bash
    ln -s /usr/local/Cellar/openssl/{openssl_version}/bin/openssl /usr/local/bin/openssl
    ```

    This will tell Mac OS to use the homebrew version of OpenSSL which includes the header and necessary development files

4.  Update `~/.bash_profile` or `~/.bashrc`
    ```bash
    export PKG_CONFIG_PATH="$(brew --prefix openssl)/lib/pkgconfig"
    ```
