package templates

import (
	"github.com/pkg/errors"

	"github.com/BurntSushi/toml"
)

type TemplateConfig struct {
	MinFabricVersion string `json:"min_fabric_version" toml:"min_fabric_version"`
}

// ParseConfig will read the provided toml file and return a configuration object
func parseConfig(path string) (TemplateConfig, error) {
	var c TemplateConfig
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return TemplateConfig{}, errors.Wrapf(err, "Failed to parse config: %s", path)
	}

	return c, nil
}
