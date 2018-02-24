package httputil

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

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

type testItem struct {
	a int
	b string
}

func TestGRPCArrayWriter(t *testing.T) {
	tests := []struct {
		name        string
		text        []string
		expectArray bool
	}{
		{name: "no data", text: nil, expectArray: false},
		{name: "not items", text: []string{`{"xxx":[]}`}, expectArray: false},
		{name: "empty array 1", text: []string{`{"items":[]}`}, expectArray: true},
		{name: "empty array 2", text: []string{`[]`}, expectArray: true},
		{name: "single item 1",
			text:        []string{`{"items":[{"a": 42,"b": "x"}]}`},
			expectArray: true,
		},
		{name: "single item 2",
			text:        []string{`[{"a": 42,"b": "x"}]`},
			expectArray: true,
		},
		{name: "multiple writes 1",
			text: []string{`{"items":[{"a": 42,"b": "x"}`,
				`]}`,
			},
			expectArray: true,
		},
		{name: "multiple writes 2",
			text: []string{`[{"a": 42,"b": "x"}`,
				`]`,
			},
			expectArray: true,
		},
		{name: "multiple items 1",
			text: []string{`{"items":[{"a": 1,"b": "a"},`,
				`{"a": 2,"b": "b"},`,
				`{"a": 3,"b": "c"},`,
				`{"a": 4,"b": "d"},`,
				`{"a": 5,"b": "e"},`,
				`{"a": 6,"b": "f"},`,
				`{"a": 7,"b": "g"},`,
				`{"a": 8,"b": "h"},`,
				`{"a": 9,"b": "i"},`,
				`{"a": 10,"b": "j"}`,
				`]}`,
			},
			expectArray: true,
		},
		{name: "multiple items 2",
			text: []string{`[{"a": 1,"b": "a"},`,
				`{"a": 2,"b": "b"},`,
				`{"a": 3,"b": "c"},`,
				`{"a": 4,"b": "d"},`,
				`{"a": 5,"b": "e"},`,
				`{"a": 6,"b": "f"},`,
				`{"a": 7,"b": "g"},`,
				`{"a": 8,"b": "h"},`,
				`{"a": 9,"b": "i"},`,
				`{"a": 10,"b": "j"}`,
				`]`,
			},
			expectArray: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
				Output(zerolog.ConsoleWriter{Out: os.Stderr})
			writer := &testWriter{t: t}
			arrayWriter := newGRPCArrayWriter(logger, writer)
			for _, textItem := range tt.text {
				if _, err := arrayWriter.Write([]byte(textItem)); err != nil {
					t.Fatalf("%s: write error: %s", tt.name, err)
				}
			}
			arrayWriter.flush()
			containsArray := writer.containsArray()
			if tt.expectArray != containsArray {
				t.Fatalf("%s: expecting array = %t, contains array = %t",
					tt.name, tt.expectArray, containsArray)
			}
		})
	}

}

// testWriter has methods that support the http.ResponseWriter interface
type testWriter struct {
	t      *testing.T
	buffer bytes.Buffer
}

func (tw *testWriter) Header() http.Header {
	tw.t.Error("unexpected call to Header()")
	return nil
}

func (tw *testWriter) Write(data []byte) (int, error) {
	tw.t.Logf("Write %d bytes", len(data))
	return tw.buffer.Write(data)
}

func (tw *testWriter) WriteHeader(status int) {
	tw.t.Errorf("unexpected call to WriteHeader(%d)", status)
}

// containsArray returns true of we can unmarshall []testItem
func (tw *testWriter) containsArray() bool {
	var array []testItem
	var err error

	data := tw.buffer.Bytes()
	if err = json.Unmarshal(data, &array); err != nil {
		tw.t.Logf("json.Unmarshal failed: %s, %s", err, string(data))
	}

	return err == nil
}
