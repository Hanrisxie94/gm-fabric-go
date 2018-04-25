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

package auth

import (
	"strings"
	"testing"
)

func TestAuth(t *testing.T) {
	type testData struct {
		name string
		dn   string
		s    string
		auth bool
	}
	for _, td := range []testData{
		// with both lists empty, we'll take anything
		{
			name: "empty lists",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": [],
				"userWhitelist": []
			}`,
			auth: true,
		},
		// if the whitelist applies to the dn, it is authorized
		{
			name: "whitelist only",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			auth: true,
		},
		// if the blacklist applies to the dn, it is unauthorized
		{
			name: "blacklist only",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": ["cn=alec.holmes,dc=deciphernow,dc=com"],
				"userWhitelist": []
			}`,
			auth: false,
		},
		// if the blacklist applies to the dn, it is unauthorized
		// even if it is on the whitelist
		{
			name: "blacklist and whitelist",
			dn:   "cn=alec.holmes,dc=deciphernow,dc=com",
			s: `{
				"userBlacklist": ["cn=alec.holmes,dc=deciphernow,dc=com"],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			auth: false,
		},
		// the whitelist does not apply to the dn, if the 'dc' RDNs are
		// in a different order
		{
			name: "RDNs out of order",
			dn:   "cn=alec.holmes,dc=com,dc=deciphernow",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["dc=deciphernow,dc=com"]
			}`,
			auth: false,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			a, err := New(strings.NewReader(td.s))
			if err != nil {
				t.Fatalf("%s: New failed: %v", td.name, err)
			}
			isAuth, err := a.IsAuthorized(td.dn)
			if err != nil {
				t.Fatalf("%s: IsAuthorized failed: %s", td.name, err)
			}
			if isAuth != td.auth {
				t.Fatalf("%s: expectedAuth = %t, result auth = %t",
					td.name, td.auth, isAuth)
			}
		})
	}
}
