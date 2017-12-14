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

package flatjson

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFlatJSONWriter(t *testing.T) {
	var buffer bytes.Buffer
	var ts map[string]interface{}
	var err error

	w, err := New(&buffer)
	if err != nil {
		t.Fatalf("New failed: %s", err)
	}
	if err = w.Write("aaa", 42); err != nil {
		t.Fatalf("WriteInt failed: %s", err)
	}
	if err = w.Write("bbb", 3.14); err != nil {
		t.Fatalf("WriteFloat32 failed: %s", err)
	}
	if err = w.Flush(); err != nil {
		t.Fatalf("Flush failed: %s", err)
	}

	data := buffer.Bytes()
	err = json.Unmarshal(data, &ts)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s; %s", err, string(data))
	}
	if ts["aaa"].(float64) != 42.0 {
		t.Fatalf("ts.aaa = %v", ts)
	}
	if ts["bbb"].(float64) != 3.14 {
		t.Fatalf("ts.bbb = %v", ts)
	}
}
