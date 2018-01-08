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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
)

// InterfaceEntry defines a parsed interface entry for the client or server
type InterfaceEntry struct {
	Comments  []string
	Prototype string
}

const (
	nullState = iota
	waitingServerState
	loadingServerState
)

// parseGeneratedPBFile the generated xxx.pb.go file returning method definitions
func parseGeneratedPBFile(
	cfg config.Config,
	logger zerolog.Logger,
) ([]InterfaceEntry, error) {
	var file *os.File
	var serverDef []InterfaceEntry
	var currentEntry InterfaceEntry
	var err error

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
	*/

	serverLine := fmt.Sprintf("type %sServer interface {", cfg.GoServiceName())
	commentLine := "\t//"
	endLine := "}"

	if file, err = os.Open(cfg.GeneratedPBFilePath()); err != nil {
		return nil, errors.Wrapf(err, "os.Open(%s)", cfg.GeneratedPBFilePath())
	}
	defer file.Close() // ignoring possible error return from Close

	state := waitingServerState
	scanner := bufio.NewScanner(file)
	for loop := true; scanner.Scan() && loop; {
		line := scanner.Text()
		switch state {
		case waitingServerState:
			if line == serverLine {
				state = loadingServerState
			}
		case loadingServerState:
			if line == endLine {
				state = nullState
				loop = false
			} else if strings.HasPrefix(line, commentLine) {
				currentEntry.Comments =
					append(currentEntry.Comments, strings.TrimSpace(line))
			} else {
				// we assume the prototype is on a single line: this is
				// generated code
				currentEntry.Prototype = addNamesToFuncDef(strings.TrimSpace(line))
				serverDef = append(serverDef, currentEntry)
				currentEntry = InterfaceEntry{}
			}
		}
	}

	return serverDef, nil
}
