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

package tlsutil

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetNormalizedDistinguishedName(t *testing.T) {
	dn := "CN=Joe Smith,OU=People,OU=development,OU=products,O=Decipher,C=US"
	apacheDN := "/C=US/O=Decipher/OU=products/OU=development/OU=People/CN=Joe Smith"

	// Normalize provided DN
	normalDN := GetNormalizedDistinguishedName(apacheDN)

	// Case in-sensitive comparison
	if strings.ToLower(normalDN) != strings.ToLower(dn) {
		t.Fatal("Normalized DN did not match expected output")
	}
}

func TestGetDNFromCert(t *testing.T) {
	trustBytes, err := ioutil.ReadFile("./testcerts/server.trust.pem")
	if err != nil {
		t.Fatal(err)
	}

	certBytes, err := ioutil.ReadFile("./testcerts/server.cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(trustBytes)
	if !ok {
		t.Fatal("failed to parse root certificate")
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		t.Fatal("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal("failed to parse certificate: " + err.Error())
	}

	dnFromCert := GetDNFromCert(cert.Issuer)
	if dnFromCert != "CN=fabric-ssl-test-server,OU=Decipher,C=US" || dnFromCert == "" {
		t.Fatal("DN from cert did not match expected output")
	}

	if strings.Index(dnFromCert, "/") != -1 {
		// assume that dn is in wrong format
		t.Fatal("DN in wrong format")
	}
}
