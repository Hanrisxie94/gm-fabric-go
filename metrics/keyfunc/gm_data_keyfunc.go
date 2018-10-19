package keyfunc

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

import (
	"net/http"
	"strings"
)

// GMDataKeyFunc is a function for defining the 'key' label in HTTP metrics
func GMDataKeyFunc(req *http.Request) string {
	if req.URL == nil {
		return ""
	}

	path := req.URL.EscapedPath()
	if len(path) == 0 {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	splitPath := strings.Split(path, "/")

	if len(splitPath) < 2 {
		return ""
	}

	return "/" + splitPath[1]
}
