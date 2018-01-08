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

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/deciphernow/gm-fabric-go/cmd/fabric/templ"
)

func generateMethod(
	cfg config.Config,
	logger zerolog.Logger,
	entry InterfaceEntry,
	methodFilePath string,
) error {
	var err error
	var template string

	methodDeclaration := entry.Prototype

	// unitary methods are of the form
	// HelloProxy(context.Context, *HelloRequest) (*HelloResponse, error)
	// stream methods are of the form
	// HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
	// So we tell them apart by looking for ' error$'
	if strings.HasSuffix(entry.Prototype, " error") {
		template = streamMethodTemplate
	} else {
		template = unitaryMethodTemplate
	}

	logger.Debug().Str("method", methodDeclaration).Msg("generateMethod")
	err = templ.Merge(
		"method",
		template,
		methodFilePath,
		struct {
			MethodsPackageName string
			PBImport           string
			Comments           string
			MethodDeclaration  string
		}{
			cfg.MethodsPackageName(),
			cfg.PBImportPath(),
			prepareComments(entry.Comments),
			methodDeclaration,
		},
	)

	if err != nil {
		return errors.Wrap(err, "templ.Merge")
	}

	return nil
}

func prepareComments(comments []string) string {
	return strings.Join(comments, "\n")
}
