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

| Package                            | Description                                                        |
|------------------------------------|--------------------------------------------------------------------|
| [consul](consul/README.md)         | Easy integration with Hashicorps service discovery system          |
| [dbutil](dbutil/README.md)         | Utilities for configuration/interaction with Mongo and Redis       |
| [events](events/README.md)         | Prototyped event streaming                                         |
| [gk](gk/README.md)                 | Gatekeeper service announcement utility                            |
| [metrics](metrics/README.md)       | GM Fabric metrics (HTTP, gRPC)                                     |
| [middleware](middleware/README.md) | HTTP middleware helpers                                            |
| [oauth](oauth/README.md)           | OAuth 2.0 authorization code (recommended use with coreos/dex)     |
| [rpcutil](rpcutil/README.md)       | Utility functions for GM Fabric gRPC services                      |
| [sds](sds/README.md)               | Service discovery and announcement (GM Fabric 2.0)                 |
| [servgen](https://github.com/DecipherNow/gm-fabric-servgen) | GM Fabric golang service generator (gRPC) |
| [tlsutil](tlsutil/README.md)       | TLS utility functions for easy integration with 2-way SSL          |
| [zkutil](zkutil/README.md)         | Helper functions for zookeeper/gatekeeper service announcement     |
| [cloudwatch](cloudwatch/README.md) | Auto-scale abilities using GM Fabric metrics and amazon cloudwatch |

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

## Install and Setup GM Fabric Go SDK

### Packages

We recommend that you use [golang/dep](https://github.com/golang/dep) for dependency management.

### Alternative

Install the repo on your local machine. There are a few options to go about this:
1.  Install with `go get`
```bash
go get -u github.com/deciphernow/gm-fabric-go
```
2.  Install with `git`
```bash
cd $GOPATH/src/github.com/deciphernow
git clone git@github.com:DecipherNow/gm-fabric-go.git
```
*Note*: if you use git, you'll want to make sure your folder structure matches the example above

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
