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

package metricsserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
)

// ReportFunc reports metrics
type ReportFunc func(*flatjson.Writer) error

// MetricsServer wraps the componenets of the metrics server
type metricsServer struct {
	reporters []ReportFunc
}

// Start starts the metrics server
// This function runs an internal goroutine
func Start(
	metricsAddress string,
	tlsConf *tls.Config,
	reporters ...ReportFunc,
) error {
	var err error
	var listener net.Listener

	ms := metricsServer{reporters: reporters}

	// decide if we should use https based on if we were given a tls config
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", metricsAddress, tlsConf)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", metricsAddress)
		if err != nil {
			return err
		}
	}

	http.HandleFunc("/metrics", ms.handleFunc)

	go http.Serve(listener, nil)

	return nil
}

func (ms metricsServer) handleFunc(w http.ResponseWriter, req *http.Request) {
	headers := w.Header()
	headers.Add("content-type", "application/json")

	jWriter, err := flatjson.New(w)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("flatjson.New failed: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	for n, reporter := range ms.reporters {
		if err = reporter(jWriter); err != nil {
			http.Error(
				w,
				fmt.Sprintf("reporter.Report failed: #%d; %v", n, err),
				http.StatusInternalServerError,
			)
			return
		}
	}

	if err = jWriter.Flush(); err != nil {
		http.Error(
			w,
			fmt.Sprintf("flatjson.Flush failed: %v", err),
			http.StatusInternalServerError,
		)
		return
	}
}
