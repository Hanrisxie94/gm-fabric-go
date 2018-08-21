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

package auth

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// rdn is a relative distinguished name (RDN). A key value pair
type rdn struct {
	key   string
	value string
}

// rdnMap represents a (partial) DN which may have multiple values for a
// key. e.g. 'dc=decipher, dc=com'. Order is significant within a key
type rdnMap map[string][]string

// AppliesTo represents the state m1 is a whitelist or blacklist entry
// and all the keys in m1 match the RDN in a DN (m2)
// m2 may have keys that are not in m1
func (m1 rdnMap) AppliesTo(m2 rdnMap) bool {
	for key, m1Slice := range m1 {
		m2Slice := m2[key]
		if len(m2Slice) != len(m1Slice) {
			return false
		}
		for i := 0; i < len(m1Slice); i++ {
			if m2Slice[i] != m1Slice[i] {
				return false
			}
		}
	}

	return true
}

// Authorizor uses white and black lists to authorize a DN
// blacklist (bl) overrides whitelist(wl)
type Authorizor struct {
	bl []rdnMap
	wl []rdnMap
}

// rawData is parsed from JSON
type rawData struct {
	UserBlacklist []string `json:"userBlackList"`
	UserWhitelist []string `json:"userWhiteList"`
}

// New creates an authorizer from JSON
func New(rawJSON io.Reader) (Authorizor, error) {
	r, err := loadRawData(rawJSON)
	if err != nil {
		return Authorizor{}, errors.Wrap(err, "loadRawData")
	}

	return NewFromLists(r.UserBlacklist, r.UserWhitelist)
}

// NewFromLists creates an authorizer from a blacklist and a whitelist
func NewFromLists(userBlacklist, userWhitelist []string) (Authorizor, error) {
	var a Authorizor
	var err error

	a.bl, err = parseList(userBlacklist)
	if err != nil {
		return Authorizor{}, errors.Wrap(err, "parseList userBlacklist")
	}

	a.wl, err = parseList(userWhitelist)
	if err != nil {
		return Authorizor{}, errors.Wrap(err, "parseList userWhitelist")
	}

	return a, nil
}

// IsAuthorized takes a DN of the form
// "cn=alec.holmes,dc=deciphernow,dc=com"
// and returns true if the DN is authorized
func (a Authorizor) IsAuthorized(rawDN string) (bool, error) {
	dnMap, err := parseDN(rawDN)
	if err != nil {
		return false, errors.Wrap(err, "parseDN")
	}

	// blacklist takes precedence of whitelist
	for _, blMap := range a.bl {
		if blMap.AppliesTo(dnMap) {
			return false, nil
		}
	}

	// if the whitelist is not empty, and they're not in the whitelist,
	// they are not authorized

	var authorized bool

	// if the whitelist is empty, pass through
	if len(a.wl) == 0 {
		authorized = true
	} else {
	WHITELIST_LOOP:
		for _, wlMap := range a.wl {
			if wlMap.AppliesTo(dnMap) {
				authorized = true
				break WHITELIST_LOOP
			}
		}
	}

	return authorized, nil
}

// loadRawData loads a JSON file of the form
// {
//		"userBlackList": [...],
//		"userWhiteList": [...]
// }
func loadRawData(rawJSON io.Reader) (rawData, error) {
	var r rawData
	dec := json.NewDecoder(rawJSON)
	err := dec.Decode(&r)
	if err != nil {
		return rawData{}, errors.Wrap(err, "Decode")
	}
	return r, nil
}

// parseList parses the content of userBlackList or userWhiteList
// It expects a list of DNs or partial DNs of the form
// [
//		"cn=alec.holmes,dc=deciphernow,dc=com",
//		...
// ]
func parseList(l []string) ([]rdnMap, error) {
	var rdnMaps []rdnMap

	for _, entry := range l {
		m, err := parseDN(entry)
		if err != nil {
			return nil, errors.Wrapf(err, "parseRDNMap: '%s'", entry)
		}
		rdnMaps = append(rdnMaps, m)
	}

	return rdnMaps, nil
}

// parseDN parses a single DN (or partial DN) of the form
// "cn=alec.holmes,dc=deciphernow,dc=com"
// returning a map of keys and possible multiple values
func parseDN(entry string) (rdnMap, error) {
	result := make(rdnMap)

	rdns, err := parseRDNs(entry)
	if err != nil {
		return nil, errors.Wrapf(err, "parseRDNs: %v", entry)
	}
	for _, rdn := range rdns {
		entry := result[rdn.key]
		result[rdn.key] = append(entry, rdn.value)
	}

	return result, nil
}

// parseRDNs parses the individual key value pairs withing a single DN
// it returns a list of pairs which will be compressed into a map
func parseRDNs(entry string) ([]rdn, error) {
	var rdns []rdn

	// split a string into a list of 'key=value' substrings
	for _, kvs := range strings.Split(entry, ",") {
		l := strings.Split(strings.TrimSpace(kvs), "=")
		if len(l) != 2 {
			return nil, errors.Errorf("unparseable kv pair: '%v'", kvs)
		}

		r := rdn{
			key:   strings.ToUpper(strings.TrimSpace(l[0])),
			value: strings.ToUpper(strings.TrimSpace(l[1])),
		}
		if r.key == "" || r.value == "" {
			return nil, errors.Errorf("key or value is blank: %+v", r)
		}

		rdns = append(rdns, r)
	}

	return rdns, nil
}
