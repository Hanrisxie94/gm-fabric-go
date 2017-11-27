# Middleware
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/middleware)

Package middleware provides a simple middleware abstraction on top of net/http.

1. [Prerequisites](#prerequisites)
2. [Install](#install)
3. [Usage](#usage)

## Prerequisites

1. [Go](https://golang.org) (1.7+ required)

## Installation

Using [golang/dep](https://github,com/golang/dep):
```bash
dep ensure -v -add github.com/deciphernow/gm-fabric-go/middleware
```

## Usage
A contrived example:
```go
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
```

When requesting GET /greetings, the response body would be as follows:
```
Entering middleware A
Entering middleware B
Hello!
Leaving middleware B
Leaving middleware A
```
