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

var zkTemplate = `package main
import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/deciphernow/gm-fabric-go/gk"
)

type zkCancelFunc func()

func notifyZkOfMetricsIfNeeded(logger zerolog.Logger) []zkCancelFunc {
	if !viper.GetBool("use_zk") {
		return nil
	}

	logger.Info().Str("service", "{{.ServiceName}}").Msg("announcing metrics endpoint to zookeeper")
	cancel := gk.Announce(viper.GetStringSlice("zk_connection_string"), &gk.Registration{
		Path:   viper.GetString("zk_announce_path") + viper.GetString("metrics_uri_path"),
		Host:   viper.GetString("zk_announce_host"),
		Status: gk.Alive,
		Port:   viper.GetInt("metrics_service_port"),
	})
	logger.Info().Str("service", "{{.ServiceName}}").Msg("Service successfully registered metrics endpoint to zookeeper")

	return []zkCancelFunc{cancel}
}

func notifyZkOfRPCServerIfNeeded(logger zerolog.Logger) []zkCancelFunc {
	if !viper.GetBool("use_zk") {
		return nil
	}

	logger.Info().Str("service", "{{.ServiceName}}").Msg("announcing rpc endpoint to zookeeper")
	cancel := gk.Announce(viper.GetStringSlice("zk_connection_string"), &gk.Registration{
		Path:   viper.GetString("zk_announce_path") + "/rpc",
		Host:   viper.GetString("zk_announce_host"),
		Status: gk.Alive,
		Port:   viper.GetInt("metrics_service_port"),
	})
	logger.Info().Str("service", "{{.ServiceName}}").Msg("Service successfully registered rpc endpoint to zookeeper")

	return []zkCancelFunc{cancel}
}

func notifyZkOfGatewayEndpointIfNeeded(logger zerolog.Logger) []zkCancelFunc {
	if !(viper.GetBool("use_zk") && viper.GetBool("use_gateway_proxy")) {
		return nil
	}

	gatewayEndpoint := "http"
	if viper.GetBool("use_tls") {
		gatewayEndpoint = "https"
	}

	logger.Info().Str("service", "{{.ServiceName}}").Msg("announcing gateway endpoint to zookeeper")

	cancel := gk.Announce(viper.GetStringSlice("zk_connection_string"), &gk.Registration{
		Path:   viper.GetString("zk_announce_path") + "/" + gatewayEndpoint,
		Host:   viper.GetString("zk_announce_host"),
		Status: gk.Alive,
		Port:   viper.GetInt("gateway_proxy_port"),
	})
	logger.Info().Str("service", "{{.ServiceName}}").Msg("announcing gateway endpoint to zookeeper")

	return []zkCancelFunc{cancel}
}
`
