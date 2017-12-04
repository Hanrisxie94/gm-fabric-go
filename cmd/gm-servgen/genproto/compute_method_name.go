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

package genproto

import (
	"strings"
	"text/scanner"
	"unicode"
)

func computeMethodFileName(prototype string) string {
	// HelloProxy(context.Context, *HelloRequest) (*HelloResponse, error)
	// convert 'HelloProxy' to 'hello_proxy.go'

	var out []rune
	var s scanner.Scanner

	s.Init(strings.NewReader(prototype))

SCAN_LOOP:
	for r := s.Next(); r != scanner.EOF; r = s.Next() {
		if r == '(' {
			break SCAN_LOOP
		}
		if unicode.IsUpper(r) {
			if len(out) > 0 {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
	}
	out = append(out, []rune(".go")...)

	return string(out)
}
