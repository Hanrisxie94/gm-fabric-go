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
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/deciphernow/gm-fabric-go/oauth"
	"github.com/deciphernow/gm-fabric-go/oauth/example/config"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

var (
	configPath = flag.String("config", "./config.toml", "path to configuration file")
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// parse the command line flags and add some seed data
	flag.Parse()
	seedData(log)

	// parse the config file
	conf := config.ParseConfig(*configPath, log)
	log.Info().Msg("Using config:")
	err := config.PrintJSON(os.Stdout, conf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to print config")
	}

	// create our HTTP router
	mux := mux.NewRouter()

	// Add our methods and their appropriate http.HandlerFunc's
	mux.HandleFunc("/movies", getMovies).Methods("GET")
	mux.HandleFunc("/movies", createMovie).Methods("POST")

	// Create our middleware stack that wraps all HTTP handlers
	stack := middleware.Chain(
		// Setup HTTP logger using github.com/rs/zerolog
		middleware.MiddlewareFunc(cors.AllowAll().Handler),
		middleware.MiddlewareFunc(hlog.NewHandler(log)),
		middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("path", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("Access")
		})),
		// Inject the GM Fabric OAuth middleware with appropriate options
		// A provider and client ID are required for proper token validation
		oauth.HTTPAuthenticate(oauth.WithProvider(conf.Oauth.Provider), oauth.WithClientID(conf.Oauth.ClientID)),
	)

	// Create our HTTP server
	s := http.Server{
		Addr: conf.Address,
		// Wrap our HTTP router with our middleware stack
		Handler: stack.Wrap(mux),
	}

	// Start the HTTP server
	log.Info().Str("address", conf.Address).Msg("Server listening. . .")
	s.ListenAndServe()
}
