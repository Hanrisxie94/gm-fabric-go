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

package initsvc

var buildRPMTemplate = `
#!/bin/sh

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC="$( cd "$DIR/.." && pwd )"

IMAGE_ID=$(docker build -q $SRC/pkg | cut -d: -f 2 -)
IMAGE_ID=${IMAGE_ID:0:12}

docker run --rm -i -v $SRC:/src "$IMAGE_ID" /bin/bash -s < $SRC/pkg/script
`
