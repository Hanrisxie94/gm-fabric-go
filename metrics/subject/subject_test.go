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

package subject

import "testing"

func TestSplitTag(t *testing.T) {
	for i, td := range []struct {
		tag      string
		expected []string
	}{
		{
			tag:      "",
			expected: []string{"", ""},
		},
		{
			tag:      "aaa",
			expected: []string{"aaa", ""},
		},
		{
			tag:      "aaa:bbb",
			expected: []string{"aaa", "bbb"},
		},
		{
			tag:      "aaa:bbb:ccc",
			expected: []string{"aaa", "bbb:ccc"},
		},
	} {
		name, value := SplitTag(td.tag)
		if name != td.expected[0] || value != td.expected[1] {
			t.Fatalf("#%d: tag mistmatch: [%v %v] != %v",
				i+1, name, value, td.expected)
		}
	}

}

func TestJoinTag(t *testing.T) {
	for i, td := range []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "",
			value:    "",
			expected: "",
		},
		{
			name:     "aaa",
			value:    "",
			expected: "aaa",
		},
		{
			name:     "aaa",
			value:    "bbb",
			expected: "aaa:bbb",
		},
		{
			name:     "aaa",
			value:    "bbb:ccc",
			expected: "aaa:bbb:ccc",
		},
	} {
		tag := JoinTag(td.name, td.value)
		if tag != td.expected {
			t.Fatalf("#%d: tag mistmatch: %v != %v",
				i+1, tag, td.expected)
		}
	}

}
