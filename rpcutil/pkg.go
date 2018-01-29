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
Package rpcutil provides helper functions for easy transition from the standard library "net/http"

An example of fetching a user_dn

		func (s *serverData) CreateCategory(ctx context.Context, request *pb.Category) (*pb.Category, error) {
			dn, err := rpcutil.FetchDNFromContext(ctx)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}

This will extract the a users distinguished name from the incoming metadata context on a grpc method.
*/
package rpcutil

import (
	"context"
	"net/textproto"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// MatchHTTPHeaders will pass through all http headers into the grpc-metadata map
func MatchHTTPHeaders(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	return key, true
}

// FetchDNFromContext will retrieve a user_dn from the incoming metadata stuffed in the gRPC method context
func FetchDNFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Errorf("rpcutil: %s", "failed to retrieve rpc metadta from incoming context")
	}

	// Find user dn in metadata. If it doesn't exist, return and fail
	if len(md["user_dn"]) == 0 {
		return "", errors.Errorf("rpcutil: %s", "failed to find user dn in context metadata")
	}
	// We get the 0 position because there will only be 1 user dn
	return md["user_dn"][0], nil
}
