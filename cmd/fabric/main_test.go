package main

import (
	"fmt"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	var testCases = []struct {
		in       string
		expected int
	}{
		{"--version", 0},
	}

	for _, tc := range testCases {
		os.Args = []string{"xxx", tc.in}
		t.Run(fmt.Sprintf("'%s'", tc.in), func(t *testing.T) {
			out := run()
			if out != tc.expected {
				t.Fatalf("expected '%d', found '%d'", tc.expected, out)
			}
		})
	}

}
