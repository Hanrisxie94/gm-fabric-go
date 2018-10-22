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
	"net/url"
	"testing"
)

func parseURL(rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return url
}

func TestDefaultKeyfunc(t *testing.T) {
	type args struct {
		req http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty request", args: args{}, want: ""},
		{name: "simple request", args: args{req: http.Request{URL: parseURL("xxx")}}, want: "xxx"},
		{name: "leading slash", args: args{req: http.Request{URL: parseURL("/aaa")}}, want: "/aaa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultHTTPKeyFunc(&tt.args.req); got != tt.want {
				t.Errorf("DefaultHTTPKeyFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
