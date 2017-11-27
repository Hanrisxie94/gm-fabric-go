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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/deciphernow/gm-fabric-go/oauth"
	"github.com/deciphernow/gm-fabric-go/oauth/example/config"
)

type errorMessage struct {
	Message string `json:"error"`
	Time    string `json:"time"`
}

type movie struct {
	Title    string `json:"title"`
	Director string `json:"director"`
	Release  string `json:"release_date"`
}

var movies []movie

// getMovies will fetch all movies in the local mem array and return the JSON
func getMovies(w http.ResponseWriter, r *http.Request) {
	// if token validation is successful, permissions will be injected in the request context
	// we can filter data accordingly with this object
	perm := oauth.RetrievePermissions(r.Context())

	// print user object to console for demonstration
	err := config.PrintJSON(os.Stdout, perm)
	if err != nil {
		config.PrintJSON(w, errorMessage{
			Message: err.Error(),
			Time:    time.Now().Format(time.RFC3339),
		})
		return
	}

	// User encoder since it implements an io.Writer
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")

	// Send back your JSON
	w.Header().Add("content-type", "application/json")
	err = encoder.Encode(movies)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// createMovie will add a movie from a request into the local movies array
func createMovie(w http.ResponseWriter, r *http.Request) {
	// if token validation is successful, permissions will be injected in the request context
	// we can filter data accordingly with this object
	perm := oauth.RetrievePermissions(r.Context())

	// Print user object to console for demonstration
	err := config.PrintJSON(os.Stdout, perm)
	if err != nil {
		config.PrintJSON(w, errorMessage{
			Message: err.Error(),
			Time:    time.Now().Format(time.RFC3339),
		})
		return
	}

	var m movie

	// Decode the incoming request body into a movie object
	if err = json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Println("Failed to read request body: " + err.Error())
	}

	// Add the movie into the movies array in mem
	movies = append(movies, m)
}
