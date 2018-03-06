package tracecontext

import (
	"testing"
)

func TestKVPair(t *testing.T) {
	testCases := []struct {
		name           string
		stringVal      string
		expectError    bool
		expectedString string
	}{
		{name: "empty", stringVal: "", expectError: true, expectedString: ""},
		{name: "normal", stringVal: "key=value", expectError: false, expectedString: "key=value"},
		{name: "space1", stringVal: " key=value", expectError: false, expectedString: "key=value"},
		{name: "space2", stringVal: " key =value", expectError: false, expectedString: "key=value"},
		{name: "space3", stringVal: " key = value", expectError: false, expectedString: "key=value"},
		{name: "no value 1", stringVal: "key=", expectError: false, expectedString: "key"},
		{name: "no value 2", stringVal: "key", expectError: false, expectedString: "key"},
	}

	for i, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			kp1, err := ParseKVPair(tc.stringVal)
			if tc.expectError {
				if err == nil {
					t.Fatalf("%d: expecting error", i)
				}
			} else {
				if err != nil {
					t.Fatalf("%d: unexpected error: %s", i, err)
				}

				s2 := kp1.String()
				if s2 != tc.expectedString {
					t.Fatalf("%d: string mismatch: %s != %s", i, s2, tc.expectedString)
				}
			}
		})
	}
}

func TestTraceState(t *testing.T) {
	testCases := []struct {
		name           string
		stringVal      string
		expectError    bool
		expectedString string
	}{
		{name: "empty", stringVal: "", expectError: true, expectedString: ""},
		{name: "normal", stringVal: "key=value", expectError: false, expectedString: "key=value"},
		{name: "space1", stringVal: " key=value", expectError: false, expectedString: "key=value"},
		{name: "space2", stringVal: " key =value", expectError: false, expectedString: "key=value"},
		{name: "space3", stringVal: " key = value", expectError: false, expectedString: "key=value"},
		{name: "no value 1", stringVal: "key=", expectError: false, expectedString: "key"},
		{name: "no value 2", stringVal: "key", expectError: false, expectedString: "key"},
		{name: "no properties", stringVal: "key=value;", expectError: true, expectedString: ""},
		{name: "single property", stringVal: "key=value;v1", expectError: false, expectedString: "key=value;v1"},
		{name: "multi property", stringVal: "key=value;k1=v1;k2=v2", expectError: false, expectedString: "key=value;k1=v1;k2=v2"},
	}

	for i, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ts1, err := ParseTraceState(tc.stringVal)
			if tc.expectError {
				if err == nil {
					t.Fatalf("%d: expecting error", i)
				}
			} else {
				if err != nil {
					t.Fatalf("%d: unexpected error: %s", i, err)
				}

				s2 := ts1.String()
				if s2 != tc.expectedString {
					t.Fatalf("%d: string mismatch: %s != %s", i, s2, tc.expectedString)
				}
			}
		})
	}
}
