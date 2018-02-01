package templates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	t "text/template"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// The callback function for invocations of the templates.Fetch method. The set of templates yielded to this method
// represent the name (i.e., relative path as a template) and content of a fetched file. This enables file names and
// paths to be treated as templates in addition to thier contents.
type Callback func(name *t.Template, content *t.Template, mode os.FileMode) error

// Fetches templates from a URL using hashicorp/go-getter and yields each file from the fetched template to the callback.
//
// Note that when invoking on a Git repository, you should only run this against a sub-directory. Running this on the root
// of a Git repository will result in the .git directory contents being yielded as templates and that is not desireable.
func Fetch(source string, callback Callback, logger zerolog.Logger) error {

	var err error
	var tmp string
	var destination string
	var templates string

	if tmp, err = ioutil.TempDir("", "fabric-"); err != nil {
		return errors.Wrapf(err, "error creating temporary directory for templates from %s", source)
	}

	defer os.RemoveAll(tmp)

	destination = path.Join(tmp, "repository")
	// TODO: read and unmarshall destination/fabric.json
	templates = path.Join(destination, "templates")

	var walker = func(absolute string, info os.FileInfo, err error) error {

		var bytes []byte
		var relative string
		var name *t.Template
		var content *t.Template

		if err != nil {
			return errors.Wrapf(err, "error walking the template directory %s", templates)
		}

		if relative, err = filepath.Rel(templates, absolute); err != nil {
			return errors.Wrapf(err, "error determining the relative path from %s to %s", templates, absolute)
		}

		if info.Mode().IsDir() {
			logger.Debug().Msg(fmt.Sprintf("Ignoring template directory: %s", relative))
			return nil
		}

		if name, err = t.New(relative).Parse(relative); err != nil {
			return errors.Wrapf(err, "error parsing the path of %s as a template", relative)
		}

		if bytes, err = ioutil.ReadFile(absolute); err != nil {
			return errors.Wrapf(err, "error reading template from %s", absolute)
		}

		if content, err = t.New(relative).Parse(string(bytes)); err != nil {
			return errors.Wrapf(err, "error parsing the content of %s as a template", relative)
		}

		return callback(name, content, info.Mode().Perm())
	}

	if err = getter.Get(destination, source); err != nil {
		return errors.Wrapf(err, "error copying templates from %s to %s", source, destination)
	}

	if err = filepath.Walk(templates, walker); err != nil {
		return errors.Wrapf(err, "error invoking callback for fetcher: %s", templates)
	}

	return nil

}
