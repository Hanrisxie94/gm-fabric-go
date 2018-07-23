package impersonation

import (
	"strings"
	"testing"
)

func TestNewWhitelist(t *testing.T) {
	dns := getServers()

	w := NewWhitelist(dns)
	for i, dn := range w.Servers {
		if strings.Compare(dns[i], dn) != 0 {
			t.Errorf("Wanted %s, got: %s", dns[i], dn)
		}
	}
}

func TestCanImpersonate(t *testing.T) {
	// Create our test whitelist
	w := NewWhitelist(getServers())
	caller := GetCaller("cn=alec.holmes,dc=deciphernow,dc=com", "uid=server3,ou=Server,dc=example,dc=com", nil)

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

func getServers() []string {
	return []string{
		"uid=server1,ou=Server,dc=example,dc=com",
		"uid=server2,ou=Server,dc=example,dc=com",
		"uid=server3,ou=Server,dc=example,dc=com",
	}
}

func BenchmarkCanImpersonate(b *testing.B) {
	// Create our test whitelist
	w := NewWhitelist(getServers())
	caller := GetCaller("cn=alec.holmes,dc=deciphernow,dc=com", "uid=server3,ou=Server,dc=example,dc=com", nil)

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
