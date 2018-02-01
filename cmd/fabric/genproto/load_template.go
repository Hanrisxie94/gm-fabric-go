package genproto

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
)

func loadTemplateFromCache(
	cfg config.Config,
	logger zerolog.Logger,
	templateName string,
) (string, error) {
	var err error
	var data []byte

	templatePath := filepath.Join(cfg.TemplateCachePath(), templateName)

	if data, err = ioutil.ReadFile(templatePath); err != nil {
		return "", errors.Wrapf(err, "error reading template from %s", templatePath)
	}

	return string(data), nil
}
