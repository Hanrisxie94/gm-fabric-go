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

package events

import (
	"encoding/json"
	"testing"
)

func TestEventAndPublisher(t *testing.T) {
	e := event{"Kaylee", 26}
	p := &publisher{}

	var success bool
	p.Publish(e)
	for _, published := range p.data {
		if ev, ok := published.(event); ok {
			if ev.Name == "Kaylee" {
				success = true
			}
		}
	}
	if !success {
		t.Fail()
	}
}

// Example types implement the Event and Publisher interfaces.

type event struct {
	Name string
	Age  int
}

func (e event) Yield() []byte {
	data, _ := json.Marshal(e)
	return data
}

type publisher struct {
	data []Event
}

func (p *publisher) Publish(e Event) {
	p.data = append(p.data, e)
}
