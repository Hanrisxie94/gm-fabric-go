package tracecontext

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// TraceStateHeaderName is the name used in an HTTP header
const TraceStateHeaderName = "Trace-State"

// TraceState is used to pass the name-value context properties for the trace.
// This is a companion header for the Trace-Parent.
type TraceState struct {
	KVPair
	Properties []KVPair
}

// KVPair represents a Key/Value pair
//Key and Value are expected to be unescaped
type KVPair struct {
	Key   string
	Value string
}

// ParseKVPair returns a KVPair from a string
//
func ParseKVPair(s string) (KVPair, error) {
	if len(s) == 0 {
		return KVPair{}, errors.Errorf("attempt to parse empty string")
	}

	var kp KVPair

	splitStr := strings.Split(s, "=")
	kp.Key = strings.TrimSpace(splitStr[0])
	if kp.Key == "" {
		return KVPair{}, errors.Errorf("Invalid blank key")
	}
	switch len(splitStr) {
	case 1:
	case 2:
		kp.Value = strings.TrimSpace(splitStr[1])
	default:
		return KVPair{}, errors.Errorf(
			"invalid format expecting 0 or 1 '=' found %d",
			len(splitStr)-1,
		)
	}

	return kp, nil
}

// String returns a formatted string of the form K=V
func (kp KVPair) String() string {
	if kp.Value == "" {
		return kp.Key
	}
	return fmt.Sprintf("%s=%s", kp.Key, kp.Value)
}

// String returns a formatted string of the TraceState
func (ts TraceState) String() string {
	if len(ts.Properties) == 0 {
		return ts.KVPair.String()
	}

	var propStr []string

	for _, prop := range ts.Properties {
		propStr = append(propStr, prop.String())
	}
	return fmt.Sprintf("%s;%s",
		ts.KVPair, strings.Join(propStr, ";"))
}

// ParseTraceState returns a TraceState from a string
func ParseTraceState(s string) (TraceState, error) {
	var state TraceState
	var err error

	kvs := strings.Split(s, ";")
	state.KVPair, err = ParseKVPair(kvs[0])
	if err != nil {
		return TraceState{}, errors.Wrap(err, "ParseKVPair")
	}

	for _, propStr := range kvs[1:] {
		prop, err := ParseKVPair(propStr)
		if err != nil {
			return TraceState{}, errors.Wrap(err, "ParseKVPair")
		}
		state.Properties = append(state.Properties, prop)
	}

	return state, nil
}
