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

import "testing"

func Test_addNamesToFuncDef(t *testing.T) {
	type args struct {
		rawDef string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{rawDef: "Hello(context.Context, *HelloRequest) (*HelloResponse, error)"},
			want: "Hello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error)",
		},
		{
			name: "simple-empty",
			args: args{rawDef: "Hello(context.Context, *empty.Empty) (*HelloResponse, error)"},
			want: "Hello(ctx context.Context, request *empty.Empty) (*pb.HelloResponse, error)",
		},
		{
			name: "stream",
			args: args{rawDef: "HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error"},
			want: "HelloStream(request *pb.HelloStreamRequest, stream pb.TestService_HelloStreamServer) error",
		},
		{
			name: "stream-empty",
			args: args{rawDef: "HelloStream(*empty.Empty, TestService_HelloStreamServer) error"},
			want: "HelloStream(request *empty.Empty, stream pb.TestService_HelloStreamServer) error",
		},
		{
			name: "unparseable",
			args: args{rawDef: "Hello(int, context.Context, *HelloRequest) (*HelloResponse, error)"},
			want: "Hello(int, context.Context, *HelloRequest) (*HelloResponse, error)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addNamesToFuncDef(tt.args.rawDef); got != tt.want {
				t.Errorf("addNamesToFuncDef() = %v, want %v", got, tt.want)
			}
		})
	}
}
