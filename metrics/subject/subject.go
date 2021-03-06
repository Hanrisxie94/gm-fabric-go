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

import (
	"context"
	"strings"
	"time"
)

// EventTransport is the scheme of the URI that initiated the event
type EventTransport uint8

const (
	// EventTransportNull should never occur
	EventTransportNull EventTransport = iota

	// EventTransportRPC from gRPC
	EventTransportRPC

	// EventTransportRPCWithTLS from gRPC
	EventTransportRPCWithTLS

	// EventTransportHTTP HTTP
	EventTransportHTTP

	// EventTransportHTTPS HTTPS
	EventTransportHTTPS
)

const TagSep = ":"

// MetricsEvent is a low level event.
type MetricsEvent struct {
	EventType  string
	Transport  EventTransport
	HTTPStatus int
	RequestID  string
	Key        string
	PrevRoute  string
	Timestamp  time.Time
	Value      interface{}
	Tags       []string
}

// Observer implements an individual observer of the observer design pattern.
// It observes an ordered stream of metrics events
type Observer interface {
	Observe(MetricsEvent)
}

// New creates a new metrics subject for feeding events to observers
func New(
	ctx context.Context,
	observers ...Observer,
) chan<- MetricsEvent {
	metricsChan := make(chan MetricsEvent)

	go func() {
		for loop := true; loop; {
			select {
			case <-ctx.Done():
				loop = false
			case event := <-metricsChan:
				for _, observer := range observers {
					observer.Observe(event)
				}
			}
		}
	}()

	return metricsChan
}

// SplitTag takes a tag of the form <name:value> and returns (name, value)
// If the tag is not splitable SplitTag returns (tag, "")
func SplitTag(tag string) (string, string) {
	result := strings.SplitN(tag, TagSep, 2)
	if len(result) == 2 {
		return result[0], result[1]
	}

	return tag, ""
}

// JoinTag joins name and value into a tag string
// Note that if name contains TagSep, the results will be unexpected
func JoinTag(name, value string) string {
	if name == "" {
		return value
	}
	if value == "" {
		return name
	}
	return strings.Join([]string{name, value}, TagSep)
}
