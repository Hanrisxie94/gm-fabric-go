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

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templates"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// InitService initializes the service up to the point where it needs a
// protocol buffer definiton
func InitService(cfg config.Config, logger zerolog.Logger) error {

	var err error
	var data interface{}

	data = struct {
		ServiceName        string
		ServicePath        string
		GoServiceName      string
		ProtoServiceName   string
		ConfigPackage      string
		ConfigPackageName  string
		MethodsPackage     string
		MethodsPackageName string
		ProtoDirName       string
		PBImport           string
		GrpcServerHost     string
		GrpcServerPort     string
		MetricsServerHost  string
		MetricsServerPort  string
		MetricsCacheSize   string
		MetricsURIPath     string
		UseGatewayProxy    bool
		GatewayProxyHost   string
		GatewayProxyPort   string
		UseTLS             bool
		CaCertPath         string
		ServerCertPath     string
		ServerKeyPath      string
		ServerCertName     string
		ReportStatsd       bool
		StatsdHost         string
		StatsdPort         string
		StatsdMemInterval  string
		VerboseLogging     bool
		UseOauth           bool
		OauthProvider      string
		OauthClientID      string
		UseZK              bool
		ZKConnectionString string
		ZKAnnouncePath     string
		ZKAnnounceHost     string
	}{
		cfg.ServiceName,
		cfg.ServicePath(),
		cfg.GoServiceName(),
		cfg.ProtoServiceName(),
		cfg.ConfigPackageImportPath(),
		cfg.ConfigPackageName(),
		cfg.MethodsImportPath(),
		cfg.MethodsPackageName(),
		cfg.ProtoDirName(),
		cfg.PBImportPath(),
		viper.GetString("grpc_server_host"),
		viper.GetString("grpc_server_port"),
		viper.GetString("metrics_server_host"),
		viper.GetString("metrics_server_port"),
		viper.GetString("metrics_cache_size"),
		viper.GetString("metrics_uri_path"),
		viper.GetBool("use_gateway_proxy"),
		viper.GetString("gateway_proxy_host"),
		viper.GetString("gateway_proxy_port"),
		viper.GetBool("use_tls"),
		viper.GetString("ca_cert_path"),
		viper.GetString("server_cert_path"),
		viper.GetString("server_key_path"),
		viper.GetString("server_cert_name"),
		viper.GetBool("report_statsd"),
		viper.GetString("statsd_host"),
		viper.GetString("statsd_port"),
		viper.GetString("statsd_mem_interval"),
		viper.GetBool("verbose_logging"),
		viper.GetBool("use_oauth"),
		viper.GetString("oauth_provider"),
		viper.GetString("oauth_client_id"),
		viper.GetBool("use_zk"),
		viper.GetString("zk_connection_string"),
		viper.GetString("zk_announce_path"),
		viper.GetString("zk_announce_host"),
	}

	logger.Info().Str("service", cfg.ServiceName).Msg("starting --init")

	var output []byte
	var callback = func(name *template.Template, content *template.Template, mode os.FileMode) error {
		if err = templates.Render(name, content, mode, cfg.ServicePath(), data, logger); err != nil {
			return errors.Wrapf(err, "Failed to render template for %s", name.Name())
		}
		return nil
	}

	logger.Debug().Msg(fmt.Sprintf("Fetching template from %s", cfg.TemplateUrl))

	if err = templates.Fetch(cfg.TemplateUrl, callback, logger); err != nil {
		return errors.Wrapf(err, "Failed to fetch template from %s", cfg.TemplateUrl)
	}

	if err = within(cfg.ServicePath(), func() error {
		if output, err = exec.Command("dep", "init").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		if output, err = exec.Command("dep", "ensure").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	if err = within(path.Join(cfg.VendorPath(), "github.com", "golang", "protobuf", cfg.ProtocGenGoPluginName()), func() error {
		logger.Debug().Msg("Installing Golang Plugin...")
		if output, err = exec.Command("go", "install", "-v").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	if err = within(path.Join(cfg.VendorPath(), "github.com", "grpc-ecosystem", "grpc-gateway", cfg.ProtocGenGatewayPluginName()), func() error {
		logger.Debug().Msg("Installing Gateway Plugin...")
		if output, err = exec.Command("go", "install", "-v").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "Failed executing command with output %", string(output))
		}
		return nil
	}); err != nil {
		return err
	}

	return nil

}

func within(directory string, callback func() error) error {

	var err error
	var current string

	if current, err = os.Getwd(); err != nil {
		return errors.Wrap(err, "Failed discover current working directory")
	}

	if err = os.Chdir(directory); err != nil {
		return errors.Wrapf(err, "Failed to change working directory to %s", directory)
	}

	defer func() {
		if err = os.Chdir(current); err != nil {
			panic(err)
		}
	}()

	if err = callback(); err != nil {
		return errors.Wrapf(err, "Failed executing callback from working directory %s", directory)
	}

	return nil
}
