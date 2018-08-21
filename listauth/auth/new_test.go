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

// TestCreateAuthNew tests creating an authorizor with New
func TestCreateAuthNew(t *testing.T) {
	type testData struct {
		name        string
		s           string
		expectError bool
	}
	for _, td := range []testData{
		{
			name:        "empty string",
			expectError: true,
		},
		{
			name: "blacklist only",
			s: `{
				"userBlacklist": []
			}`,
			expectError: false,
		},
		{
			name: "whitelist only",
			s: `{
				"userWhitelist": []
			}`,
			expectError: false,
		},
		{
			name: "both lists empty",
			s: `{
				"userBlacklist": [],
				"userWhitelist": []
			}`,
			expectError: false,
		},
		{
			name: "invalid blacklist content",
			s: `{
				"userBlacklist": ["jjj"],
				"userWhitelist": []
			}`,
			expectError: true,
		},
		{
			name: "valid blacklist content",
			s: `{
				"userBlacklist": ["key=value"],
				"userWhitelist": []
			}`,
			expectError: false,
		},
		{
			name: "invalid whitelist content",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["xxx"]
			}`,
			expectError: true,
		},
		{
			name: "valid whitelist content",
			s: `{
				"userBlacklist": [],
				"userWhitelist": ["key1=value1", "key2=value2"]
			}`,
			expectError: false,
		},
		{
			name: "valid both list content",
			s: `{
				"userBlacklist": ["key0=value0"],
				"userWhitelist": [
					"key1=value1, key2=value2",
					"key5=value5",
					"key6=value6, key7=value7"
				]
			}`,
			expectError: false,
		},
	} {
		// run the tests with JSON input
		t.Run(td.name, func(t *testing.T) {
			_, err := New(strings.NewReader(td.s))
			foundError := err != nil
			if foundError != td.expectError {
				t.Fatalf("%s: expectedError = %t, foundError = %t",
					td.name, td.expectError, foundError)
			}
		})
	}
}

// TestCreateAuthNewFromLists tests creating an authorizor with NewFromLists
func TestCreateAuthNewFromLists(t *testing.T) {
	type testData struct {
		name        string
		blacklist   []string
		whitelist   []string
		expectError bool
	}
	for _, td := range []testData{
		{
			name:        "both lists empty",
			expectError: false,
		},
		{
			name:        "invalid blacklist content",
			blacklist:   []string{"jjj"},
			expectError: true,
		},
		{
			name:        "valid blacklist content",
			blacklist:   []string{"key=value"},
			expectError: false,
		},
		{
			name:        "invalid whitelist content",
			whitelist:   []string{"xxx"},
			expectError: true,
		},
		{
			name:        "valid whitelist content",
			whitelist:   []string{"key1=value1", "key2=value2"},
			expectError: false,
		},
		{
			name:      "valid both list content",
			blacklist: []string{"key0=value0"},
			whitelist: []string{
				"key1=value1, key2=value2",
				"key5=value5",
				"key6=value6, key7=value7",
			},
			expectError: false,
		},
	} {
		// run the tests with list input
		t.Run(td.name, func(t *testing.T) {
			_, err := NewFromLists(td.blacklist, td.whitelist)
			foundError := err != nil
			if foundError != td.expectError {
				t.Fatalf("%s: expectedError = %t, foundError = %t",
					td.name, td.expectError, foundError)
			}
		})
	}
}
