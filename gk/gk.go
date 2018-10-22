// Copyright 2018 Decipher Technology Studios LLC
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

package gk

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/zkutil"

	"github.com/samuel/go-zookeeper/zk"
)

// -----------------------------------------------------------------------------
// The JSON blob stored in ZooKeeper.
// These correspond to the following parts of the GateKeeper config:
//
//   zookeeper.jsonPath.host=serviceEndpoint.host
//   zookeeper.jsonPath.port=serviceEndpoint.port
//
// Here's an example of the JSON:
//
//	{
//		"serviceEndpoint": {
//			"host": "127.0.0.1",
//			"port": 8080
//		},
//		"status": "ALIVE"
//	}
//
// TODO: The JSON schema should be handled dynamically, such that one can
// specify the jsonPath to use, rather than baking that into the AnnounceData
// type.

// AnnounceData models the data written to a ZooKeeper ephemeral node.
type AnnounceData struct {
	ServiceEndpoint Address `json:"serviceEndpoint"`
	Status          status  `json:"status"`
}

// Address models a host + port combination.
type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// don't export; force the user to use one of the constants
type status string

// Alive refers to current service status in zk
const Alive = status("ALIVE")

// GetIP returns an IPv4 Address in string format suitable for Gatekeeper to reach us at
func GetIP() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	if len(hostname) <= 0 {
		return "", errors.New("could not find our hostname")
	}
	myIPs, err := net.LookupIP(hostname)
	if err != nil {
		myIPs, err = net.LookupIP("localhost")
		if err != nil {
			return "", errors.New("could not get a set of IPs for our hostname")
		}
	}
	if len(myIPs) <= 0 {
		return "", errors.New("could not find IPv4 address")
	}
	for a := range myIPs {
		if myIPs[a].To4() != nil {
			return myIPs[a].String(), nil
		}
	}
	return "", errors.New("could not find IPv4 address")
}

// Registration struct holds the necessary info to perform a proper announcement to zookeeper
type Registration struct {
	Path   string
	Status status
	Host   string
	Port   int
}

func (r *Registration) toAnn() AnnounceData {
	json := AnnounceData{
		Status: r.Status,
		ServiceEndpoint: Address{
			Host: r.Host,
			Port: r.Port,
		},
	}
	return json
}

// Announce Registers the service announcement with ZooKeeper.
// Should look something like:
//
//	cancel := gk.Announce([]string{"localhost:2181"}, &gk.Registration{
//		Path:   "/cte/service/category/1.0.0",
//		Host:   "127.0.0.1",
//		Port:   8090,
//		Status: gk.Alive,
//	})
//	defer cancel()
//
// This function is deprecated, it is kept for compatibility
func Announce(servers []string, reg *Registration, opts ...Opt) (cancel func()) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return AnnounceWithLogger(servers, reg, logger, opts...)
}

// AnnounceWithLogger registers the service announcement with ZooKeeper.
// Should look something like:
//
//	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
//		Output(zerolog.ConsoleWriter{Out: os.Stderr})
//
// serviceLogger := logger.With().Str("service", "<service-name>").Logger()
//
//	cancel := gk.AnnounceWithLogger(
// 		[]string{"localhost:2181"},
//		&gk.Registration{
//			Path:   "/cte/service/category/1.0.0",
//			Host:   "127.0.0.1",
//			Port:   8090,
//			Status: gk.Alive,
//		},
//		serviceLogger.
// )
//	defer cancel()
func AnnounceWithLogger(servers []string, reg *Registration, logger zerolog.Logger, opts ...Opt) (cancel func()) {
	done := make(chan struct{})
	cancel = func() {
		close(done)
	}

	annJSON := reg.toAnn()
	annBytes, _ := json.Marshal(annJSON)

	go func() {
		// Announce until cancelled.
		for doAnn(done, annBytes, servers, reg, logger, opts...) {
			select {
			case <-done:
				return
			case <-time.After(time.Second * 2):
				// Retry.
			}
		}
	}()

	return cancel
}

type zkZeroLogger struct {
	logger zerolog.Logger
}

func (zl zkZeroLogger) Printf(format string, a ...interface{}) {
	zl.logger.Debug().Msgf(format, a...)
}

func doAnn(done chan struct{}, annBytes []byte, servers []string, reg *Registration, logger zerolog.Logger, opts ...Opt) bool {
	// Set are options
	var options Options
	for _, o := range opts {
		o(&options)
	}

	zl := zkZeroLogger{logger: logger}
	var conn *zk.Conn
	var err error
	var evCh <-chan zk.Event
	if options.TLS != nil {
		conn, evCh, err = connectWithTLS(servers, 2*time.Second, zl, options.TLS)
		if err != nil {
			logger.Error().AnErr("zk.Connect", err).Msg("failed to connect with TLS")
			// Time to reconnect.
			return true
		}
	} else {
		conn, evCh, err = zk.Connect(servers, 2*time.Second, zk.WithLogger(zl))
		if err != nil {
			logger.Error().AnErr("zk.Connect", err).Msg("")
			// Time to reconnect.
			return true
		}
	}

	defer conn.Close()

	expired := true

create:
	_, err = zkutil.CreateRecursive(conn, reg.Path+"/member_", annBytes, zk.FlagEphemeral|zk.FlagSequence, zkutil.DefaultACLs)
	if err == nil {
		// Mark that we successfully created the node.
		expired = false
	} else {
		logger.Error().AnErr("zkutil.CreateRecursive", err).Msg("")
	}

	// Wait until we're cancelled or the connection fails.
	for {
		select {
		case <-done:
			// Bail.
			return false
		case ev := <-evCh:
			// The zk.Conn will attempt to reconnect repeatedly upon disconnect,
			// and as long as a connection is established within the session timeout,
			// the ephemeral nodes will continue to exist.
			// On the other hand, if the session expires, we need to recreate the
			// node.
			switch {
			case ev.Err != nil:
				logger.Error().AnErr("Gatekeeper announcement", err).Msg("")
				return true
			case ev.State == zk.StateExpired:
				logger.Info().Msg("Gatekeeper announcement expired")
				expired = true
			case expired && ev.State == zk.StateHasSession:
				logger.Info().Msg("Re-announcing service")
				goto create
			}
		}
	}
}

// connectWithTLS will create a zookeeper connection with a TLS config
func connectWithTLS(servers []string, t time.Duration, zl zkZeroLogger, cfg *tls.Config) (*zk.Conn, <-chan zk.Event, error) {
	conn, evCh, err := zk.Connect(servers, 2*time.Second, zk.WithLogger(zl), zk.WithDialer(zk.Dialer(func(network string, address string, timeout time.Duration) (net.Conn, error) {
		// Establish a TCP connection to the address with the specified timeout
		// using the DialTimeout method
		ipConn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			log.Printf("Could not connect to %v, %v\n", address, network)
			return ipConn, err
		}
		log.Printf("TCP Connected to %v, %v\n", address, network)

		// Use the TCP connection created above to establish the TLS connection
		// Need to use the Client method since we already have the TCP connection
		conn := tls.Client(ipConn, cfg)
		err = conn.Handshake()
		if err != nil {
			return nil, err
		}

		return conn, nil
	})))
	if err != nil {
		zl.logger.Error().AnErr("zk.Connect", err).Msg("")
		// Time to reconnect.
		return nil, nil, err
	}

	return conn, evCh, nil
}
