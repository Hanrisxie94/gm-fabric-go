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

package apistats

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
)

func TestReport(t *testing.T) {
	const key = "xxx"
	entry := APIStatsEntry{Key: key, BeginTime: time.Now(), EndTime: time.Now()}

	stats := New(1)
	stats.Store(entry)

	var buffer bytes.Buffer

	w, err := flatjson.New(&buffer)
	if err != nil {
		t.Fatalf("New failed: %s", err)
	}

	err = stats.Report(w)
	if err != nil {
		t.Fatalf("stats.Report failed: %s", err)
	}
	err = w.Flush()
	if err != nil {
		t.Fatalf("w.Flush() failed: %s", err)
	}

	var ts map[string]interface{}
	data := buffer.Bytes()
	err = json.Unmarshal(data, &ts)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s; %s", err, string(data))
	}

	t.Logf("ts = %q", ts)

}
