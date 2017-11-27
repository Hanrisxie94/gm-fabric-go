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

package dbutil

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestCreateHash(t *testing.T) {
	hash := CreateHash()
	if hash == "" {
		t.FailNow()
	}
}

func TestWriteJSON(t *testing.T) {
	var message struct {
		M string `json:"message"`
	}
	message.M = "Testing WriteJSON..."

	err := WriteJSON(os.Stdout, message)
	if err != nil {
		t.FailNow()
	}
}

func TestReadRequest(t *testing.T) {
	var body struct {
		Message string `json:"message"`
	}
	body.Message = "Hello World!"

	b, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	req, err := http.NewRequest("GET", "http://testing.test", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	// clear out body message in mem
	body.Message = ""

	// Test our read req function
	err = ReadReqest(req, &body)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if body.Message == "" {
		t.FailNow()
	}
}
