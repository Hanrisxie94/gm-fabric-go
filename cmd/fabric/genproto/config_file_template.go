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

package genproto

// settings.toml
var configFileTemplate = `
# grpc-server
    grpc_server_host = "{{.GrpcServerHost}}"
    grpc_server_port = {{.GrpcServerPort}}

# metrics-server
    metrics_server_host =  "{{.MetricsServerHost}}"
    metrics_server_port = {{.MetricsServerPort}}
    metrics_cache_size =  {{.MetricsCacheSize}}
    metrics_uri_path = "{{.MetricsURIPath}}"

# gateway-proxy
    use_gateway_proxy = "{{.UseGatewayProxy}}"
    gateway_proxy_host = "{{.GatewayProxyHost}}"
    gateway_proxy_port = {{.GatewayProxyPort}}

# tls
    use_tls = {{.UseTLS}}
    ca_cert_path = "{{.CaCertPath}}"
    server_cert_path = "{{.ServerCertPath}}"
    server_key_path = "{{.ServerKeyPath}}"
    server_cert_name = "{{.ServerCertName}}"

# statsd
    report_statsd = {{.ReportStatsd}}
    statsd_host = "{{.StatsdHost}}"
    statsd_port = {{.StatsdPort}}
    statsd_mem_interval = "{{.StatsdMemInterval}}"

# misc
    verbose_logging = true
`
