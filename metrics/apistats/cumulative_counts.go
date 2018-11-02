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

import "github.com/deciphernow/gm-fabric-go/metrics/subject"

type KeyEventsEntry struct {
	Events            int64
	StatusEvents      map[int]int64
	StatusClassEvents map[string]int64
}

type CumulativeCounts struct {
	TotalEvents     int64
	TransportEvents map[subject.EventTransport]int64
	KeyEvents       map[string]KeyEventsEntry
}

func newCumulativeCounts() CumulativeCounts {
	return CumulativeCounts{
		TransportEvents: make(map[subject.EventTransport]int64),
		KeyEvents:       make(map[string]KeyEventsEntry),
	}
}

func copyCumulativeCounts(inp CumulativeCounts) CumulativeCounts {
	outp := newCumulativeCounts()
	outp.TotalEvents = inp.TotalEvents
	for key, value := range inp.KeyEvents {
		outp.KeyEvents[key] = value
	}
	for key, value := range inp.TransportEvents {
		outp.TransportEvents[key] = value
	}

	return outp
}
