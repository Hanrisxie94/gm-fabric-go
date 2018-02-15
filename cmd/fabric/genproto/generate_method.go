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
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templates"
)

func generateMethod(
	cfg config.Config,
	logger zerolog.Logger,
	entry PrototypeEntry,
	methodFilePath string,
) error {
	var err error
	var templateName string
	var templ string
	var protobufImport string
	var pbImport string

	methodDeclaration := entry.Prototype

	// unitary methods are of the form
	// HelloProxy(context.Context, *HelloRequest) (*HelloResponse, error)
	// stream methods are of the form
	// HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
	// So we tell them apart by looking for ' error$'
	if strings.HasSuffix(entry.Prototype, " error") {
		templateName = "stream_method.go"
	} else {
		templateName = "unitary_method.go"
	}
	templ, err = loadTemplateFromCache(cfg, logger, templateName)
	if err != nil {
		return errors.Wrapf(err, "loadTemplateFromCache %s", templateName)
	}
	if strings.Contains(methodDeclaration, "google_protobuf") {
		protobufImport = `google_protobuf "github.com/golang/protobuf/ptypes/empty"`
	}
	if strings.Contains(methodDeclaration, "*pb.") {
		pbImport = fmt.Sprintf(`pb "%s"`, cfg.PBImportPath())
	}

	logger.Debug().Str("method", methodDeclaration).Msg("generateMethod")

	err = templates.Merge(
		"method",
		templ,
		methodFilePath,
		struct {
			MethodsPackageName string
			ProtobufImport     string
			PBImport           string
			Comments           string
			MethodDeclaration  string
		}{
			cfg.MethodsPackageName(),
			protobufImport,
			pbImport,
			prepareComments(entry.Comments),
			methodDeclaration,
		},
	)

	if err != nil {
		return errors.Wrap(err, "templ.Merge")
	}

	cmd := exec.Command(
		"gofmt",
		"-w",
		methodFilePath,
	)
	if op, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "gofmt -w %s: %s", methodFilePath, string(op))
	}

	return nil
}

func prepareComments(comments []string) string {
	return strings.Join(comments, "\n")
}
