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

package config

import (
	"fmt"
	"testing"
)

func TestConvertToCamelCase(t *testing.T) {
	var testCases = []struct {
		in       string
		expected string
	}{
		{"aaa", "Aaa"},
		{"Aaa", "Aaa"},
		{"aaa_bbb", "AaaBbb"},
		{"Aaa_bbb", "AaaBbb"},
		{"Aaa_Bbb", "AaaBbb"},
		{"AaaBbb", "AaaBbb"},
		{"Odrive2text", "Odrive2Text"},
		{"Odrive2Text", "Odrive2Text"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("'%s'", tc.in), func(t *testing.T) {
			out := convertToCamelCase(tc.in)
			if out != tc.expected {
				t.Fatalf("expected '%s', found '%s'", tc.expected, out)
			}
		})
	}

}
