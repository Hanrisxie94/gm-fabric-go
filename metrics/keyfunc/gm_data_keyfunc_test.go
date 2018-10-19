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
	"testing"
)

func TestGMDataKeyFunc(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty request", args: args{}, want: ""},
		{name: "simple request", args: args{"xxx"}, want: "/xxx"},
		{name: "case from Issue #176",
			args: args{"/notifications/0000000000000001/?last=155e7732bb79d53c&count=10000"},
			want: "/notifications",
		},
		{name: "/stream/${oid}/${path}",
			args: args{"/stream/1/home/robpike/gopher.png"},
			want: "/stream",
		},
		{name: "/props/${oid}/${path}",
			args: args{"/props/1/home/robpike/gopher.png"},
			want: "/props",
		},
		{name: "/list/${oid}/${path}",
			args: args{"/list/1/home/robpike/gopher.png"},
			want: "/list",
		},
		{name: "/history/${oid}/${path}",
			args: args{"/history/1/home/robpike/gopher.png"},
			want: "/history",
		},
		{name: "/notifications/${oid}/${path}",
			args: args{"/notifications/1/"},
			want: "/notifications",
		},
		{name: "/derived/${oid}/${path}",
			args: args{"/derived/4321432/"},
			want: "/derived",
		},
		{name: "/html/${oid}/${path}",
			args: args{"/html/4321432/"},
			want: "/html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{URL: parseURL(tt.args.rawURL)}
			if got := GMDataKeyFunc(&req); got != tt.want {
				t.Errorf("GMDataKeyFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
