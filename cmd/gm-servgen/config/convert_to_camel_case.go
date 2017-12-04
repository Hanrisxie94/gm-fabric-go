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

package config

import (
	"strings"
	"unicode"
)

func convertToCamelCase(inStr string) string {
	n := strings.Replace(inStr, "_", " ", -1)
	n = strings.Title(n)
	n = strings.Replace(n, " ", "", -1)

	// Make any rune preceded by a number upper case
	// get prevRune from the closure so it persists between calls
	var prevRune rune
	mapFunc := func(r rune) rune {
		result := r
		if prevRune != 0 && unicode.IsLetter(r) && unicode.IsNumber(prevRune) {
			result = unicode.ToUpper(r)
		}
		prevRune = r
		return result
	}

	return strings.Map(mapFunc, n)
}
