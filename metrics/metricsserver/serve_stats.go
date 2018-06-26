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
	"log"
	"net/http"
)

// Start starts the metrics server
// This function runs an internal goroutine
// This function is deprecated: retained for compatibility
func Start(
	metricsAddress string,
	tlsConf *tls.Config,
	reporters ...ReportFunc,
) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", NewDashboardHandler(reporters...))

	server := NewMetricsServer(metricsAddress, tlsConf)
	server.Handler = mux

	if server.TLSConfig == nil {
		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Printf("server.ListenAndServe() failed: %v", err)
			}
		}()
	} else {
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil {
				log.Printf("server.ListenAndServeTLS() failed: %v", err)
			}
		}()
	}

	return nil
}
