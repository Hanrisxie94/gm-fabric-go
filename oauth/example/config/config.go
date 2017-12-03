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

package config

import (
	"encoding/json"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
)

type oauth struct {
	ClientID string `json:"oauth_client_id" toml:"clientID"`
	Provider string `json:"oauth_provider" toml:"provider"`
}

// Config - basic global config
type Config struct {
	Oauth   oauth  `json:"oauth" toml:"oauth"`
	Address string `json:"address" toml:"address"`
}

// ParseConfig will read the provided toml file and return a global configuration object
func ParseConfig(path string, log zerolog.Logger) *Config {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		log.Error().Err(err).Msg("Failed to parse config")
	}

	return &c
}

// PrintJSON will pretty print JSON using an IO.reader to whatever output source it is provided
func PrintJSON(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(value)
	if err != nil {
		return err
	}
	return nil
}
