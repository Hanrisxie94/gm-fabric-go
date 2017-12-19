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
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/version"
)

// Operation requested on command line
type Operation int8

const (
	// Null is an invalid Operation
	Null Operation = iota

	// ShowVersion means display version string to stdout and exit
	ShowVersion

	// Init is initialize a new service
	Init

	// Generate protocol buffer code for an existing service
	Generate
)

// Config contains program configuration
type Config struct {
	Op          Operation
	Version     string
	ServiceName string
	OwnerDir    string
	LogLevel    zerolog.Level
}

// Load constructs the configuation from various sources
func Load() (Config, error) {
	var showVersion bool
	var initService string
	var generateService string
	var configFilePath string
	var logLevel string
	var cfg Config
	var err error

	pflag.BoolVar(&showVersion, "version", false,
		"display version string to stdout and exit")
	pflag.StringVar(&initService, "init", "",
		"initialize service")
	pflag.StringVar(&generateService, "generate", "",
		"generate protobuff for service")
	pflag.StringVar(&cfg.OwnerDir, "dir", "",
		"path to the directory containing the service. Default: cwd")
	pflag.StringVar(&configFilePath, "config", "",
		"path to the fabric_settings.toml file. Default: no file")
	pflag.StringVar(&logLevel, "log-level", "error",
		"global log level ('debug', 'info', 'error'): default 'error'")
	pflag.Parse()

	cfg.Version = version.Current()

	switch strings.ToLower(logLevel) {
	case "debug":
		cfg.LogLevel = zerolog.DebugLevel
	case "info":
		cfg.LogLevel = zerolog.InfoLevel
	case "error":
		cfg.LogLevel = zerolog.ErrorLevel
	default:
		return Config{}, fmt.Errorf("Unknown log level '%s'", logLevel)
	}

	switch {
	case showVersion:
		cfg.Op = ShowVersion
		return cfg, nil
	case initService != "" && generateService != "":
		return Config{}, errors.New("cannot specify both --init and --generate")
	case initService != "":
		cfg.Op = Init
		cfg.ServiceName = initService
	case generateService != "":
		cfg.Op = Generate
		cfg.ServiceName = generateService
	default:
		return Config{}, errors.New("you must specify either --init or --generate")
	}

	if cfg.OwnerDir == "" {
		if cfg.OwnerDir, err = os.Getwd(); err != nil {
			return Config{}, errors.Wrap(err, "os.Getwd")
		}
	}

	viper.SetDefault("grpc_server_port", 10000)
	viper.SetDefault("metrics_server_port", 10001)
	viper.SetDefault("metrics_cache_size", 1024)
	viper.SetDefault("metrics_uri_path", "/metrics")
	viper.SetDefault("gateway_proxy_port", 8080)
	viper.SetDefault("statsd_host", "127.0.0.1")
	viper.SetDefault("statsd_port", 8125)

	if configFilePath != "" {
		viper.SetConfigName(strings.Split(filepath.Base(configFilePath), ".")[0])
		viper.AddConfigPath(filepath.Dir(configFilePath))
	} else {
		viper.SetConfigName("fabric_settings")
		viper.AddConfigPath("/etc/fabric/")
		viper.AddConfigPath("$HOME/.fabric")
		viper.AddConfigPath(".")
	}

	// we don't care if ReadInConfig returns an error: we'll run off the defaults
	viper.ReadInConfig()

	return cfg, nil
}

// ServicePath the path to the top service directory
func (cfg Config) ServicePath() string {
	return path.Join(cfg.OwnerDir, cfg.ServiceName)
}

// DockerDirName the path to the directory where docker files reside
func (cfg Config) DockerDirName() string {
	return "docker"
}

// DockerPath the path to the directory where docker files reside
func (cfg Config) DockerPath() string {
	return path.Join(cfg.ServicePath(), cfg.DockerDirName())
}

// DockerEntryPointPath the path to the docker entrypoint.sh file
func (cfg Config) DockerEntryPointPath() string {
	return path.Join(cfg.DockerPath(), "entrypoint.sh")
}

// BuildDockerImageScriptName the name of the bash file to build docker image
func (cfg Config) BuildDockerImageScriptName() string {
	return fmt.Sprintf("build_%s_docker_image.sh", cfg.ServiceName)
}

// BuildDockerImageScriptPath the name to the bash file to Build docker image
func (cfg Config) BuildDockerImageScriptPath() string {
	return path.Join(cfg.ServicePath(), cfg.BuildDockerImageScriptName())
}

// RunDockerImageScriptName the name of the bash file to run the docker image
func (cfg Config) RunDockerImageScriptName() string {
	return fmt.Sprintf("run_%s_docker_image.sh", cfg.ServiceName)
}

// RunDockerImageScriptPath the name to the bash file to run the docker image
func (cfg Config) RunDockerImageScriptPath() string {
	return path.Join(cfg.ServicePath(), cfg.RunDockerImageScriptName())
}

// ProtoDirName the path to the directory where protocol buffer files reside
func (cfg Config) ProtoDirName() string {
	return "protobuf"
}

// ProtoPath the path to the directory where protocol buffer files reside
func (cfg Config) ProtoPath() string {
	return path.Join(cfg.ServicePath(), cfg.ProtoDirName())
}

// PBImportPath the path to import for access to generated protocol buffer code
func (cfg Config) PBImportPath() string {
	pbImport := strings.TrimPrefix(cfg.ProtoPath(), gopathSrc())
	return strings.TrimLeft(pbImport, string(filepath.Separator))
}

// CmdPath the path to the 'cmd' directory
func (cfg Config) CmdPath() string {
	return path.Join(cfg.ServicePath(), "cmd")
}

// VendorPath the path to the vendor directory
func (cfg Config) VendorPath() string {
	return path.Join(cfg.ServicePath(), "vendor")
}

// SettingsFilePath path to the generated settings.toml file
func (cfg Config) SettingsFilePath() string {
	return path.Join(cfg.ServicePath(), "settings.toml")
}

// BuildServerScriptName the name of the bash file to build server
func (cfg Config) BuildServerScriptName() string {
	return fmt.Sprintf("build_%s_server.sh", cfg.ServiceName)
}

// BuildServerScriptPath the name to the bash file to build server
func (cfg Config) BuildServerScriptPath() string {
	return path.Join(cfg.ServicePath(), cfg.BuildServerScriptName())
}

// GRPCClientPath the path to the 'cmd/grpc_client' directory
func (cfg Config) GRPCClientPath() string {
	return path.Join(cfg.CmdPath(), "grpc_client")
}

// BuildGRPCClientScriptName the name of the bash file to Build grpc client
func (cfg Config) BuildGRPCClientScriptName() string {
	return fmt.Sprintf("build_%s_grpc_client.sh", cfg.ServiceName)
}

// BuildGRPCClientScriptPath the name to the bash file to Build grpc client
func (cfg Config) BuildGRPCClientScriptPath() string {
	return path.Join(cfg.ServicePath(), cfg.BuildGRPCClientScriptName())
}

// ServerPath the path to the 'cmd/server' directory
func (cfg Config) ServerPath() string {
	return path.Join(cfg.CmdPath(), "server")
}

// ServerGatewayProxySourceFilePath the path to 'cmd/server/gateway_proxy.go'
func (cfg Config) ServerGatewayProxySourceFilePath() string {
	return path.Join(cfg.ServerPath(), "gateway_proxy.go")
}

// ConfigPackageName the path to the directory where the config package resides
func (cfg Config) ConfigPackageName() string {
	return "config"
}

// ConfigPackagePath the path to the directory where config package resides
func (cfg Config) ConfigPackagePath() string {
	return path.Join(cfg.ServerPath(), cfg.ConfigPackageName())
}

// ConfigPackageImportPath the path to import the config package
func (cfg Config) ConfigPackageImportPath() string {
	cpkgImport := strings.TrimPrefix(cfg.ConfigPackagePath(), gopathSrc())
	return strings.TrimLeft(cpkgImport, string(filepath.Separator))
}

// MethodsPackageName the name of the methods package
func (cfg Config) MethodsPackageName() string {
	return "methods"
}

// MethodsPath the path to the methods package
func (cfg Config) MethodsPath() string {
	return path.Join(cfg.ServerPath(), cfg.MethodsPackageName())
}

// MethodsImportPath the path to import the methods package
func (cfg Config) MethodsImportPath() string {
	mi := strings.TrimPrefix(cfg.MethodsPath(), gopathSrc())
	return strings.TrimLeft(mi, string(filepath.Separator))
}

// ProtoServiceName is service name modified for protocol buffers
// hyphen replaced with underscore
func (cfg Config) ProtoServiceName() string {
	return strings.Replace(cfg.ServiceName, "-", "_", -1)
}

// GoServiceName is service name modified by protoc for go
// "Names are turned from camel_case to CamelCase for export.""
func (cfg Config) GoServiceName() string {
	return convertToCamelCase(cfg.ProtoServiceName())
}

// ProtocGenGoPluginName is the name of the plugin for --go_out=plugins=grpc:
func (cfg Config) ProtocGenGoPluginName() string {
	return "protoc-gen-go"
}

// ProtocGenGoPluginPath is the path to the plugin for --go_out=plugins=grpc:
func (cfg Config) ProtocGenGoPluginPath() string {
	return filepath.Join(gopathBin(), cfg.ProtocGenGoPluginName())
}

// ProtocGenGatewayPluginName is the name of the plugin for --grpc-gateway_out
func (cfg Config) ProtocGenGatewayPluginName() string {
	return "protoc-gen-grpc-gateway"
}

// ProtocGenGatewayPluginPath is the path to the plugin for --grpc-gateway_out
func (cfg Config) ProtocGenGatewayPluginPath() string {
	return filepath.Join(gopathBin(), cfg.ProtocGenGatewayPluginName())
}

// ProtoFileName the name of the protocol buffer file
func (cfg Config) ProtoFileName() string {
	return fmt.Sprintf("%s.proto", strings.ToLower(cfg.ProtoServiceName()))
}

// ProtoFilePath full path to the the protocol buffer file
func (cfg Config) ProtoFilePath() string {
	return path.Join(cfg.ProtoPath(), cfg.ProtoFileName())
}

// GeneratedPBFileName the name of the generated gRPC go file
func (cfg Config) GeneratedPBFileName() string {
	return fmt.Sprintf("%s.pb.go", strings.ToLower(cfg.ProtoServiceName()))
}

// GeneratedPBFilePath full path to <service-name>.pb.go
func (cfg Config) GeneratedPBFilePath() string {
	return path.Join(cfg.ProtoPath(), cfg.GeneratedPBFileName())
}

// GeneratedPBProxyName the name of the generated gateway proxy file
func (cfg Config) GeneratedPBProxyName() string {
	return fmt.Sprintf("%s.pb.gw.go", strings.ToLower(cfg.ProtoServiceName()))
}

// GeneratedPBProxyPath full path to <service-name>.pb.gw.go
func (cfg Config) GeneratedPBProxyPath() string {
	return path.Join(cfg.ProtoPath(), cfg.GeneratedPBProxyName())
}

// GitIgnorePath full path to .gitignore in root of service directory
func (cfg Config) GitIgnorePath() string {
	return path.Join(cfg.ServicePath(), ".gitignore")
}

// RPMBundlingPath is the full path to the /rpm dir which olds artifacts for bundling
func (cfg Config) RPMBundlingPath() string {
	return path.Join(cfg.ServicePath(), "/rpm")
}

func gopath() string {
	path := os.Getenv("GOPATH")
	if path == "" {
		path = filepath.Join(os.Getenv("HOME"), "go")
	}
	return path
}

func gopathSrc() string {
	return filepath.Join(gopath(), "src")
}

func gopathBin() string {
	return filepath.Join(gopath(), "bin")
}
