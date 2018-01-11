package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func Render(name *template.Template, content *template.Template, mode os.FileMode, destination string, data interface {}, logger zerolog.Logger) error {

	var buffer bytes.Buffer
	var err error
	var file *os.File
	var absolute string
	var parent string
	var relative string

	if err = name.Execute(&buffer, data); err != nil {
		return errors.Wrapf(err, "Failed to render name template for %s", name.Name())
	} 

	relative = buffer.String()
	absolute = filepath.Join(destination, relative)
	parent = filepath.Join(destination, filepath.Dir(relative))

	if err = os.MkdirAll(parent, 0777); err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", parent)
	}

	if file, err = os.Create(absolute); err != nil {
		return errors.Wrapf(err, "Failed to create file %s", relative)
	}

	defer file.Close()

	logger.Debug().Msg(fmt.Sprintf("Rendering template for %s to %s", relative, absolute))

	if err = content.Execute(file, data); err != nil {
		return errors.Wrapf(err, "Failed to render content template for %s", content.Name())
	}

	if err = os.Chmod(absolute, mode); err != nil {
		return errors.Wrapf(err, "Failed to set permissions for %s", absolute)
	}

	return nil

}
