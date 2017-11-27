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

package logobserver

import (
	"log"
	"os"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

type logobs struct {
	l *log.Logger
}

// New creates a logging observer
// if logger is nil, the default (stderr) loggr will be used
func New(logger *log.Logger) subject.Observer {
	var lo logobs
	if logger == nil {
		lo.l = log.New(os.Stderr, "", log.LstdFlags)
	} else {
		lo.l = logger
	}

	return lo
}

// Observe implements the subject.Observer interface
// Observe an event by logging it
func (lo logobs) Observe(event subject.MetricsEvent) {
	lo.l.Printf(
		"EventType=%s "+
			"RequestID=%s "+
			"Key=%s "+
			"Timestamp=%s "+
			"Value=%+v",
		event.EventType,
		event.RequestID,
		event.Key,
		event.Timestamp,
		event.Value,
	)
}
