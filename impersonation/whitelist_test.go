package impersonation

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/rs/zerolog"
)

func TestNewWhitelist(t *testing.T) {
	dns := getServers()

	w := NewWhitelist(dns)
	for i, dn := range w.servers {
		if strings.Compare(dns[i], dn) != 0 {
			t.Errorf("Wanted %s, got: %s", dns[i], dn)
		}
	}
}

func TestCanImpersonate(t *testing.T) {
	// Create our test whitelist
	w := NewWhitelist(getServers())
	caller := GetCaller("cn=alec.holmes,dc=deciphernow,dc=com", "uid=server3,ou=Server,dc=example,dc=com", "", nil)

	// Test the first case of impersonation
	ci := CanImpersonate(caller, w)
	if !ci {
		t.Error("Expected true, got false")
	}

	// Change the sys DN to something that shouldn't pass
	caller.ExternalSystemDistinguishedName = "uid=server4,ou=Server,dc=example,dc=com"
	ci = CanImpersonate(caller, w)
	if ci {
		t.Error("Expected false, got true")
	}
}

func TestValidateCaller(t *testing.T) {
	s := buildImpersonationServer()

	for _, dn := range getServers() {
		req, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			t.Error(err)
		}
		req.Header.Set("USER_DN", "cn=alec.holmes,dc=deciphernow,dc=com")
		req.Header.Set("EXTERNAL_SYS_DN", "")
		req.Header.Set("SSL_CLIENT_S_DN", dn)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Error(fmt.Errorf("Expected 200, got %d", res.StatusCode))
		}
	}
}

func BenchmarkCanImpersonate(b *testing.B) {
	// Create our test whitelist
	w := NewWhitelist(getServers())
	caller := GetCaller("cn=alec.holmes,dc=deciphernow,dc=com", "uid=server3,ou=Server,dc=example,dc=com", "", nil)

	// Test the first case of impersonation
	ci := CanImpersonate(caller, w)
	if !ci {
		b.Error("Expected true, got false")
	}

	// Change the sys DN to something that shouldn't pass
	caller.ExternalSystemDistinguishedName = "uid=server4,ou=Server,dc=example,dc=com"
	ci = CanImpersonate(caller, w)
	if ci {
		b.Error("Expected false, got true")
	}
}

func BenchmarkValidateCaller(b *testing.B) {
	s := buildImpersonationServer()

	for _, dn := range getServers() {
		req, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			b.Error(err)
		}
		req.Header.Set("USER_DN", "cn=alec.holmes,dc=deciphernow,dc=com")
		req.Header.Set("EXTERNAL_SYS_DN", dn)
		req.Header.Set("SSL_CLIENT_S_DN", dn)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Error(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			b.Error(fmt.Errorf("Expected 200, got %d", res.StatusCode))
		}
	}
}

func getServers() []string {
	return []string{
		"uid=server1,ou=Server,dc=example,dc=com",
		"uid=server2,ou=Server,dc=example,dc=com",
		"uid=server3,ou=Server,dc=example,dc=com",
	}
}

func buildImpersonationServer() *httptest.Server {
	w := NewWhitelist(getServers())

	m := middleware.Middleware(
		ValidateCaller(w, zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})),
	)

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world")
	})

	s := httptest.NewServer(m.Wrap(router))

	return s
}
