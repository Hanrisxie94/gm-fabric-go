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

var serverTemplate = `package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/gmfabricsink"
	"github.com/deciphernow/gm-fabric-go/metrics/gometricsobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcmetrics"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	ms "github.com/deciphernow/gm-fabric-go/metrics/metricsserver"
	"github.com/deciphernow/gm-fabric-go/metrics/sinkobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/deciphernow/gm-fabric-go/tlsutil"

	"{{.ConfigPackage}}"
	"{{.MethodsPackage}}"
	pb "{{.PBImport}}"

	// we don't use this directly, but need it in vendor for gateway grpc plugin
	_ "github.com/golang/glog"
	_ "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func main() {
	var tlsServerConf *tls.Config
	var err error

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger.Info().Str("service", "{{.ServiceName}}").Msg("starting")

	ctx, cancelFunc := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Debug().Str("service", "{{.ServiceName}}").Msg("initializing config")
	if err = {{.ConfigPackageName}}.Initialize(); err != nil {
		logger.Fatal().AnErr("{{.ConfigPackageName}}.Initialize()", err).Msg("")
	}

	logger.Debug().Str("service", "{{.ServiceName}}").Msg("creating server")
	server, err := methods.New{{.GoServiceName}}Server()
	if err != nil {
		logger.Fatal().AnErr("New{{.GoServiceName}}Server())", err).Msg("")
	}

	if viper.GetBool("use_tls") {
		logger.Debug().Str("service", "{{.ServiceName}}").
			Str("ca_cert_path", viper.GetString("ca_cert_path")).
			Str("server_cert_path", viper.GetString("server_cert_path")).
			Str("server_key_path", viper.GetString("server_key_path")).
			Msg("loading TLS config")
		tlsServerConf, err = tlsutil.BuildServerTLSConfig(
			viper.GetString("ca_cert_path"),
			viper.GetString("server_cert_path"),
			viper.GetString("server_key_path"),
		)
		if err != nil {
			logger.Fatal().AnErr("tlsutil.BuildServerTLSConfig", err).Msg("")
		}
	}

	logger.Debug().Str("service", "{{.ServiceName}}").
		Str("host", viper.GetString("grpc_server_host")).
		Int("port", viper.GetInt("grpc_server_port")).
		Msg("creating listener")

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			"%s:%d",
			viper.GetString("grpc_server_host"),
			viper.GetInt("grpc_server_port"),
		),
	)
	if err != nil {
		logger.Fatal().AnErr("net.Listen", err).Msg("")
	}

	grpcObserver := grpcobserver.New(viper.GetInt("metrics_cache_size"))
	goMetObserver := gometricsobserver.New()
	observers := []subject.Observer{grpcObserver, goMetObserver}

	if viper.GetBool("report_statsd") {
		var statsdSink *gometrics.StatsiteSink

		statsdSink, err = gometrics.NewStatsiteSink(
			fmt.Sprintf(
				"%s:%d",
				viper.GetString("statsd_server_host"),
				viper.GetInt("statsd_server_port"),
			),
		)
		if err != nil {
			logger.Fatal().AnErr("gometrics.NewStatsiteSink", err).Msg("")
		}
		sinkObserver := sinkobserver.New(
			statsdSink,
			viper.GetDuration("statsd_mem_interval"),
		)
		logger.Debug().Str("service", "{{.ServiceName}}").
			Str("host", viper.GetString("statsd_server_host")).
			Int("port", viper.GetInt("statsd_server_port")).
			Dur("interval", viper.GetDuration("statsd_mem_interval")).
			Msg("reporting statsd")
		observers = append(observers, sinkObserver)
	}

	logger.Debug().Str("service", "{{.ServiceName}}").
		Str("host", viper.GetString("metrics_server_host")).
		Int("port", viper.GetInt("metrics_server_port")).
		Msg("starting metrics server")
	err = ms.Start(
		fmt.Sprintf(
			"%s:%d",
			viper.GetString("metrics_server_host"),
			viper.GetInt("metrics_server_port"),
		),
		tlsServerConf,
		grpcObserver.Report,
		goMetObserver.Report,
	)
	if err != nil {
		logger.Fatal().AnErr("start metrics server", err).Msg("")
	}

	metricsChan := subject.New(ctx, observers...)

	sink := gmfabricsink.New(metricsChan)
	gometrics.NewGlobal(gometrics.DefaultConfig("{{.ServiceName}}"), sink)

	statsHandler := grpcmetrics.NewStatsHandler(metricsChan)
	opts := []grpc.ServerOption{
		grpc.StatsHandler(statsHandler),
	}
	if viper.GetBool("use_tls") {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsServerConf)))
	}

	grpcServer := grpc.NewServer(opts...)

	pb.Register{{.GoServiceName}}Server(grpcServer, server)

	logger.Debug().Str("service", "{{.ServiceName}}").
		Msg("starting grpc server")
	go grpcServer.Serve(lis)

	if viper.GetBool("use_gateway_proxy") {
		logger.Debug().Str("service", "{{.ServiceName}}").
			Msg("starting gateway proxy")
		go gatewayProxy(ctx, logger)
	}

	s := <- sigChan
	logger.Info().Str("service", "{{.ServiceName}}") .
		Str("signal", s.String()).
		Msg("shutting down")
	cancelFunc()
	grpcServer.Stop()
}
`
