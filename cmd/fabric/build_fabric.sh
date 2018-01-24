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

SRVGEN_PATH=$GOPATH/src/github.com/deciphernow/gm-fabric-go/cmd/fabric

# If cgo is explicitly disabled, then we can't use the race detection flag
GOFLAGS=-race
if [ "x$CGO_ENABLED" == "x0" ]
then
  echo Go race condition detection is disabled in this build, because you disabled CGO
  GOFLAGS=
fi

pushd $SRVGEN_PATH
GITHASH=`(git rev-parse --verify --short HEAD)`
go install $GOFLAGS \
    -ldflags "-X github.com/deciphernow/gm-fabric-go/version.gitHash=${GITHASH}"
popd
