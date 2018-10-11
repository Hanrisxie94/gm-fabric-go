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

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestConfig(t *testing.T) {
	const ownerDir = "/owner/dir"
	const serviceName = "service-name"
	const protoServiceName = "service_name"
	const goServiceName = "ServiceName"
	const protoFileName = "service_name.proto"

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	os.Args = []string{"xxx", "--init", serviceName, "--dir", ownerDir}
	cfg, err := Load(logger)
	if err != nil {
		t.Fatalf("Load failed: %s", err)
	}

	tests := []struct {
		testFunc func() string
		want     string
	}{
		{cfg.ServicePath, fmt.Sprintf("%s/%s", ownerDir, serviceName)},
		{cfg.DockerDirName, "docker"},
		{cfg.DockerPath, fmt.Sprintf("%s/%s/%s", ownerDir, serviceName, "docker")},
		{cfg.DockerEntryPointPath, fmt.Sprintf("%s/%s/%s/entrypoint.sh", ownerDir, serviceName, "docker")},
		{cfg.BuildDockerImageScriptName, fmt.Sprintf("build_%s_docker_image.sh", serviceName)},
		{cfg.BuildDockerImageScriptPath, fmt.Sprintf("%s/%s/build_%s_docker_image.sh", ownerDir, serviceName, serviceName)},
		{cfg.RunDockerImageScriptName, fmt.Sprintf("run_%s_docker_image.sh", serviceName)},
		{cfg.RunDockerImageScriptPath, fmt.Sprintf("%s/%s/run_%s_docker_image.sh", ownerDir, serviceName, serviceName)},
		{cfg.ProtoDirName, "protobuf"},
		{cfg.ProtoPath, fmt.Sprintf("%s/%s/protobuf", ownerDir, serviceName)},
		{cfg.PBImportPath, "owner/dir/service-name/protobuf"},
		{cfg.LocalDataPath, fmt.Sprintf("%s/%s/.fabric", ownerDir, serviceName)},
		{cfg.TemplateCachePath, fmt.Sprintf("%s/%s/.fabric/templates", ownerDir, serviceName)},
		{cfg.CmdPath, fmt.Sprintf("%s/%s/cmd", ownerDir, serviceName)},
		{cfg.VendorPath, fmt.Sprintf("%s/%s/vendor", ownerDir, serviceName)},
		{cfg.SettingsFilePath, fmt.Sprintf("%s/%s/settings.toml", ownerDir, serviceName)},
		{cfg.BuildServerScriptName, fmt.Sprintf("build_%s_server.sh", serviceName)},
		{cfg.BuildServerScriptPath, fmt.Sprintf("%s/%s/build_%s_server.sh", ownerDir, serviceName, serviceName)},
		{cfg.GRPCClientPath, fmt.Sprintf("%s/%s/cmd/grpc_client", ownerDir, serviceName)},
		{cfg.BuildGRPCClientScriptName, fmt.Sprintf("build_%s_grpc_client.sh", serviceName)},
		{cfg.BuildGRPCClientScriptPath, fmt.Sprintf("%s/%s/build_%s_grpc_client.sh", ownerDir, serviceName, serviceName)},
		{cfg.HTTPClientPath, fmt.Sprintf("%s/%s/cmd/http_client", ownerDir, serviceName)},
		{cfg.BuildHTTPClientScriptName, fmt.Sprintf("build_%s_http_client.sh", serviceName)},
		{cfg.BuildHTTPClientScriptPath, fmt.Sprintf("%s/%s/build_%s_http_client.sh", ownerDir, serviceName, serviceName)},
		{cfg.ServerPath, fmt.Sprintf("%s/%s/cmd/server", ownerDir, serviceName)},
		{cfg.ServerGatewayProxySourceFilePath, fmt.Sprintf("%s/%s/cmd/server/gateway_proxy.go", ownerDir, serviceName)},
		{cfg.ConfigPackageName, "config"},
		{cfg.ConfigPackagePath, fmt.Sprintf("%s/%s/cmd/server/config", ownerDir, serviceName)},
		{cfg.ConfigPackageImportPath, "owner/dir/service-name/cmd/server/config"},
		{cfg.MethodsPackageName, "methods"},
		{cfg.MethodsPath, fmt.Sprintf("%s/%s/cmd/server/methods", ownerDir, serviceName)},
		{cfg.MethodsImportPath, "owner/dir/service-name/cmd/server/methods"},
		{cfg.ProtoServiceName, protoServiceName},
		{cfg.GoServiceName, goServiceName},
		{cfg.ProtocGenGoPluginName, "protoc-gen-go"},
		{cfg.ProtocGenGoPluginPath, fmt.Sprintf("%s/protoc-gen-go", gopathBin())},
		{cfg.ProtocGenGatewayPluginName, "protoc-gen-grpc-gateway"},
		{cfg.ProtocGenGatewayPluginPath, fmt.Sprintf("%s/protoc-gen-grpc-gateway", gopathBin())},
		{cfg.ProtocGenSwaggerPluginName, "protoc-gen-swagger"},
		{cfg.ProtocGenSwaggerPluginPath, fmt.Sprintf("%s/protoc-gen-swagger", gopathBin())},
		{cfg.ProtoFileName, protoFileName},
		{cfg.ProtoFilePath, fmt.Sprintf("%s/%s/protobuf/%s", ownerDir, serviceName, protoFileName)},
		{cfg.GeneratedPBFileName, "service_name.pb.go"},
		{cfg.GeneratedPBFilePath, fmt.Sprintf("%s/%s/protobuf/service_name.pb.go", ownerDir, serviceName)},
		{cfg.GeneratedPBProxyName, "service_name.pb.gw.go"},
		{cfg.GeneratedPBProxyPath, fmt.Sprintf("%s/%s/protobuf/service_name.pb.gw.go", ownerDir, serviceName)},
		{cfg.GitIgnorePath, fmt.Sprintf("%s/%s/.gitignore", ownerDir, serviceName)},
		{cfg.RPMBundlingPath, fmt.Sprintf("%s/%s/rpm", ownerDir, serviceName)},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("#%d", i+1), func(t *testing.T) {
			got := tt.testFunc()
			if got != tt.want {
				t.Errorf("#%d got %v, want %v", i+1, got, tt.want)
			}
		})
	}
}
