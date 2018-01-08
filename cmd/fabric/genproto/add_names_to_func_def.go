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

/*
	we expect something like this
	...
	type SomethingServer interface {
		// Hello simply says 'hello' to the server
		Hello(context.Context, *HelloRequest) (*HelloResponse, error)
		// HelloProxy says 'hello' in a form that is handled by the gateway proxy
		HelloProxy(context.Context, *HelloRequest) (*HelloRequest, error)
		// HelloStream returns multiple replies
		HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
	}
	...

	we want to convert

	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	to
	Hello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error)

	and

	HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
	to
	HelloStream(request *pb.HelloStreamRequest, stream pb.TestService_HelloStreamServer) error

*/

import (
	"fmt"
	"regexp"
)

var (
	// Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	unitaryRegexp = regexp.MustCompile(`^\s*(\S+)\(context.Context,\s\*(\S+)\)\s\(\*(\S+), error\).*$`)

	// HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
	streamRegexp = regexp.MustCompile(`^\s*(\S+)\(\*(\S+),\s(\S+)\)\serror.*$`)
)

func addNamesToFuncDef(rawDef string) string {
	if matches := unitaryRegexp.FindStringSubmatch(rawDef); matches != nil {
		name := matches[1]
		request := matches[2]
		response := matches[3]
		return fmt.Sprintf(
			"%s(ctx context.Context, request *pb.%s) (*pb.%s, error)",
			name,
			request,
			response,
		)
	}
	if matches := streamRegexp.FindStringSubmatch(rawDef); matches != nil {
		name := matches[1]
		request := matches[2]
		stream := matches[3]
		return fmt.Sprintf(
			"%s(request *pb.%s, stream pb.%s) error",
			name,
			request,
			stream,
		)
	}
	return rawDef
}
