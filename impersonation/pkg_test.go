package impersonation

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/deciphernow/gm-fabric-go/tlsutil"
)

func TestGetCaller(t *testing.T) {
	const delay = time.Second * 5

	server := buildTestServer(t)
	defer server.Shutdown(context.Background())
	go func() {
		err := server.ListenAndServeTLS("./testcerts/localhost.crt", "./testcerts/localhost.key")
		t.Logf("server.ListenAndServeTLS ends with %s", err)
	}()

	t.Logf("waiting %s for server to start", delay)
	time.Sleep(delay)

	req := buildRequest(t)
	client, err := tlsutil.NewTLSClientConnFactory(
		"./testcerts/intermediate.crt",
		"./testcerts/localhost.crt",
		"./testcerts/localhost.key",
		"localhost",
		"0.0.0.0",
		"1111",
	)
	if err != nil {
		t.Error(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 403 {
		t.Error(errors.New("Failed impersonation check"))
	}
}

func buildTestServer(t *testing.T) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world")

		var cert *x509.Certificate
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			cert = r.TLS.PeerCertificates[0]
		}
		caller := GetCaller(
			r.Header.Get("USER_DN"),
			r.Header.Get("SSL_CLIENT_S_DN"),
			r.Header.Get("EXTERNAL_SYS_DN"),
			cert,
		)

		if caller.SystemDistinguishedName == "" {
			t.Error(errors.New("SystemDistinguishedName can not be empty"))
		}
	})

	cfg, err := tlsutil.NewTLSConfig(
		"./testcerts/intermediate.crt",
		"./testcerts/localhost.crt",
		"./testcerts/localhost.key",
		tlsutil.WithClientAuth(tls.RequireAndVerifyClientCert),
	)
	if err != nil {
		t.Error(err)
	}

	s := http.Server{
		TLSConfig: cfg,
		Handler:   router,
		Addr:      "0.0.0.0:1111",
	}

	return &s
}

func buildRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest("GET", "https://localhost:1111/", nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("USER_DN", "cn=alec.holmes")

	return req
}
