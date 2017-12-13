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

package gometricsobserver

import (
	"fmt"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
)

// Report implements the Reporter interface it is called by the metrics server
func (g *GoMetricsObserver) Report(jWriter *flatjson.Writer) error {
	var err error

	g.Lock()
	defer g.Unlock()

	for key, gauge := range g.GaugeMap {
		prefixKey := fmt.Sprintf("go_metrics/%s", key)
		if err = jWriter.Write(prefixKey, gauge.Value); err != nil {
			return err
		}
	}

	for key, emitKey := range g.EmitKeyMap {
		prefixKey := fmt.Sprintf("go_metrics/%s", key)
		if err = jWriter.Write(prefixKey, emitKey.Value); err != nil {
			return err
		}
	}

	for key, counter := range g.CounterMap {
		prefixKey := fmt.Sprintf("go_metrics/%s", key)
		if err = jWriter.Write(prefixKey, counter.Value); err != nil {
			return err
		}
	}

	for key, sample := range g.SampleMap {
		prefixKey := fmt.Sprintf("go_metrics/%s", key)
		if err = jWriter.Write(prefixKey, sample.Value); err != nil {
			return err
		}
	}

	return nil
}
