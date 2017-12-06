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

package templ

import (
	"os"
	"text/template"

	"github.com/pkg/errors"
)

// Merge produces an output file by merging a template with replacement data
func Merge(
	templateName string,
	templateText string,
	outputPath string,
	replacements interface{},
) error {
	var file *os.File
	var replacer *template.Template
	var err error

	if file, err = os.Create(outputPath); err != nil {
		return errors.Wrapf(err, "os.Create(%s)", outputPath)
	}
	defer file.Close() // ignoring error return

	if replacer, err = template.New(templateName).Parse(templateText); err != nil {
		return errors.Wrap(err, "template.Parse")
	}
	if err = replacer.Execute(file, replacements); err != nil {
		return errors.Wrap(err, "replacer.Execute")
	}

	return nil
}
