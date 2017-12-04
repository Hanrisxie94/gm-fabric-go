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

var buildDockerImageTemplate = `#!/bin/bash

set -euxo pipefail

pushd {{.ServerPath}}
CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o {{.DockerPath}}/{{.ServiceName}} .
popd

pushd {{.DockerPath}}
cp {{.SettingsFilePath}} .
docker build -t  {{.ServiceName}} .
popd

`
