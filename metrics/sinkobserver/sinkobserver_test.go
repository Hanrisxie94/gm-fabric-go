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

package sinkobserver

import "testing"

func Test_fixEntryKey(t *testing.T) {
	type args struct {
		rawKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid simple key",
			args: args{rawKey: "os"},
			want: "os",
		},
		{
			name: "invalid simple key",
			args: args{rawKey: "*os*"},
			want: "x_os_",
		},
		{
			name: "valid function",
			args: args{rawKey: "function/HelloStream/errors.count"},
			want: "function:HelloStream:errors_count",
		},
		{
			name: "invalid function",
			args: args{rawKey: "function/HelloStream/xxx/errors.count"},
			want: "xfunction_HelloStream_xxx_errors_count",
		},
		{
			name: "valid route",
			args: args{rawKey: "route/repos/deciphernow/bouncycastle-maven-plugin/issues/GET/latency_ms.avg"},
			want: "route:repos_deciphernow_bouncycastle_maven_plugin_issues:GET:latency_ms_avg",
		},
		{
			name: "invalid route",
			args: args{rawKey: "route/GET/latency_ms.avg"},
			want: "xroute_GET_latency_ms_avg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixEntryKey(tt.args.rawKey); got != tt.want {
				t.Errorf("fixEntryKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
