package prometheus

import (
	"net/http"
	"net/http/httptest"
	"time"

	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestHTTPWrappper(t *testing.T) {

	hf, err := NewHandlerFactory(0.0, 0.5, 10)
	if err != nil {
		t.Fatalf("NewHandlerFactory failed: %s", err)
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
