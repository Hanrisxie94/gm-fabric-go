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

var dockerFileTemplate = `FROM ubuntu

COPY ./settings.toml /etc/{{.ServiceName}}/settings.toml
COPY ./{{.ServiceName}} /usr/bin/{{.ServiceName}}
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE {{.GrpcServerPort}}
EXPOSE {{.MetricsServerPort}}
EXPOSE {{.GatewayProxyPort}}
`
