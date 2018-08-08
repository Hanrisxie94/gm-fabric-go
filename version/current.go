package version

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

import "fmt"

// semver is the semantic version
// see https://semver.org/
const semver = "0.1.10"

// gitHash should be filled in at compile time with
// GITHASH=`(git rev-parse --verify --short HEAD)`
// go install -ldflags "-X github.com/deciphernow/gm-fabric-go/version.gitHash=${GITHASH}"
var gitHash string

// Current reports the current version of the gm-fabric-go library
func Current() string {
	return fmt.Sprintf("%s (%s)", semver, gitHash)
}

// Raw returns a version suitable for parsing by github.com/hashicorp/go-version
func Raw() string {
	return semver
}
