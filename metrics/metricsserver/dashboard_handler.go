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
	"fmt"
	"net/http"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
)

// ReportFunc reports dashboard metrics
type ReportFunc func(*flatjson.Writer) error

type dashboardHandler struct {
	reporters []ReportFunc
}

// NewDashboardHandler returns an object that implments the http.Handler interface
// It returns a flat JSON map suitable for Fabric Dashboard metrics
func NewDashboardHandler(reporters ...ReportFunc) http.Handler {
	return dashboardHandler{reporters: reporters}
}

func (dh dashboardHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	for n, reporter := range dh.reporters {
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
