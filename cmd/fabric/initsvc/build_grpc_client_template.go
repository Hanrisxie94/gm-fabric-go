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

var buildGRPCClientTemplate = `#!/bin/bash

set -euxo pipefail

# assume we are in the service base directory
BASEDIR=$PWD
CLIENTDIR="${BASEDIR}/{{.ServiceName}}/cmd/grpc_client"

(
	cd $CLIENTDIR
	go build -o=$GOPATH/bin/{{.ServiceName}}_grpc_client
)

`
