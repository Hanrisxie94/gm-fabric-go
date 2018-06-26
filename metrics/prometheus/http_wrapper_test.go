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
package prometheus

import (
	"net/http"
	"net/http/httptest"
	"time"

	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestHTTPWrappper(t *testing.T) {

	hf, err := NewSummaryHandlerFactory()
	if err != nil {
		t.Fatalf("NewSummaryHandlerFactory failed: %s", err)
	}

	handler, err := hf.NewHandler(http.HandlerFunc(testHandler))
	if err != nil {
		t.Fatalf("NewHandler( failed: %s", err)
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	http.Get("/ace/hardware")

	promhttp.Handler()

}

func testHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Second)
}
