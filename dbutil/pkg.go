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

// Package dbutil implements basic utility functions that may be useful for interfacing with various databases such as Mongo and Redis
package dbutil

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/garyburd/redigo/redis"
)

type key int

const (
	redisPoolKey key = iota
	mongoSessKey
)

// NewRedisConnection will establish a redis conneciton and perform authentication if a password is provided
func NewRedisConnection(connectionString string, password string) *redis.Pool {
	// create a redis connection pool
	pool := &redis.Pool{
		// Reasonable defaults for a redis connection pool
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			var conn redis.Conn

			conn, err := redis.Dial("tcp", connectionString)
			if err != nil {
				log.Println("failed to connect to redis: " + err.Error())
				return conn, err
			}

			// if a password is provided, authenticate with redis
			if password != "" {
				_, err := conn.Do("AUTH", password)
				if err != nil {
					log.Println("failed to authenticate with redis: " + err.Error())
					return conn, err
				}
			}

			return conn, nil
		},
	}

	return pool
}

// WithRedis takes a redis connection pool, and injects a single connection into the request context
func WithRedis(pool *redis.Pool) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redisConnection := pool.Get()
			defer redisConnection.Close()

			r = r.WithContext(
				context.WithValue(r.Context(), redisPoolKey, redisConnection),
			)

			next.ServeHTTP(w, r)
		})
	})
}

// GetRedis extracts the redis connection from the context
func GetRedis(ctx context.Context) (redis.Conn, error) {
	conn, ok := ctx.Value(redisPoolKey).(redis.Conn)
	if !ok {
		return conn, errors.New("failed to extract redis connection pool from context")
	}

	return conn, nil
}

// NewMongoSession will establish a mongo connection with the provided connection string and catch any errors that may occur
func NewMongoSession(connectionString string) *mgo.Session {
	sess, err := mgo.Dial(connectionString)
	if err != nil {
		panic(err)
	}

	return sess
}

// WithMongo will pass around a mongo  session into the request context
func WithMongo(sess *mgo.Session) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// pull a connection from the pool.
			reqSess := sess.Copy()
			// Close after use to clean up
			defer reqSess.Close()

			r = r.WithContext(
				context.WithValue(r.Context(), mongoSessKey, reqSess),
			)

			next.ServeHTTP(w, r)
		})
	})
}

// GetMongo will extract a mongo session from a context
func GetMongo(ctx context.Context) (*mgo.Session, error) {
	sess, ok := ctx.Value(mongoSessKey).(*mgo.Session)
	if !ok {
		return nil, errors.New("failed to extract mongo session from context")
	}

	return sess, nil
}

// CreateHash creates a random ID of type (string) suitable for object IDs. The current time in nanoseconds is used as a seed to the pseudo-random number generators
func CreateHash() string {
	// Use current time in nanoseconds
	source := rand.NewSource(time.Now().UnixNano())

	b := make([]byte, 32)
	binary.LittleEndian.PutUint64(b[0:], rand.New(source).Uint64())

	hasher := sha256.New()
	hasher.Write(b)

	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash
}

// WriteJSON takes in a writer and interface to encode and return JSON.
func WriteJSON(w io.Writer, value interface{}) error {
	// Use an encoder since it uses a writer (async)
	encoder := json.NewEncoder(w)
	// Pretty print JSON
	encoder.SetIndent("", "    ")

	// encode provided value to the writer
	return encoder.Encode(value)
}

// ReadReqest reads out the body of the *http.Request object into the value object passed as the second parameter.
func ReadReqest(r *http.Request, value interface{}) error {
	// decode the http request body into the provided value
	err := json.NewDecoder(r.Body).Decode(value)
	if err != nil {
		return err
	}

	// if all goes smoothly, return nil
	return nil
}
