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

/*
Package middleware provides a simple middleware abstraction on top of net/http.

A contrived example:

	import (
		"fmt"
		"log"
		"net/http"

		"github.com/deciphernow/gm-fabric-go/middleware"
	)

	func main() {
		m := middleware.Chain(
			debug("A"),
			debug("B"),
		)

		http.Handle("/greetings", m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello!\n")
		})))

		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	func debug(msg string) middleware.Middleware {
		return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Entering middleware %s\n", msg)
				next.ServeHTTP(w, r)
				fmt.Fprintf(w, "Leaving middleware %s\n", msg)
			})
		})
	}

When requesting GET /greetings, the response body would be as follows:

	Entering middleware A
	Entering middleware B
	Hello!
	Leaving middleware B
	Leaving middleware A
*/
package middleware

import "net/http"

// A Middleware Wraps an existing http.Handler.
type Middleware interface {
	Wrap(http.Handler) http.Handler
}

// The MiddlewareFunc type is an adapter to allow the use of ordinary functions
// as HTTP middleware. If f is a function with the appropriate signature,
// MiddlewareFunc(f) is a Middleware that calls f.
type MiddlewareFunc func(next http.Handler) http.Handler

// Wrap calls f(next).
func (self MiddlewareFunc) Wrap(next http.Handler) http.Handler {
	return self(next)
}

/*
Compose a set of middleware into a single middleware.

Given the following:

	Chain(f, g, h)

f would be run first, followed by g and h.
*/
func Chain(ms ...Middleware) Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			next = ms[i].Wrap(next)
		}
		return next
	})
}
