package httputil

// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apackage httputil

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

import (
	"bytes"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type arrayWriter struct {
	logger      zerolog.Logger
	next        http.ResponseWriter
	isGrpcArray bool
	status      int
	prevBuf     []byte
}

// NewGRPCArrayWriter returns an object that implments the http.ResponseWriter
// interface and also a 'flush' method to force the last Write
func newGRPCArrayWriter(
	logger zerolog.Logger,
	next http.ResponseWriter,
) *arrayWriter {
	return &arrayWriter{logger: logger, next: next}
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
//
// Changing the header map after a call to WriteHeader (or
// Write) has no effect unless the modified headers are
// trailers.
//
// There are two ways to set Trailers. The preferred way is to
// predeclare in the headers which trailers you will later
// send by setting the "Trailer" header to the names of the
// trailer keys which will come later. In this case, those
// keys of the Header map are treated as if they were
// trailers. See the example. The second way, for trailer
// keys not known to the Handler until after the first Write,
// is to prefix the Header map keys with the TrailerPrefix
// constant value. See TrailerPrefix.
//
// To suppress implicit response headers (such as "Date"), set
// their value to nil.
func (aw *arrayWriter) Header() http.Header {
	return aw.next.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
//
// If WriteHeader has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data. If the Header
// does not contain a Content-Type line, Write adds a Content-Type set
// to the result of passing the initial 512 bytes of written data to
// DetectContentType.
//
// Depending on the HTTP protocol version and the client, calling
// Write or WriteHeader may prevent future reads on the
// Request.Body. For HTTP/1.x requests, handlers should read any
// needed request body data before writing the response. Once the
// headers have been flushed (due to either an explicit Flusher.Flush
// call or writing enough data to trigger a flush), the request body
// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
// handlers to continue to read the request body while concurrently
// writing the response. However, such behavior may not be supported
// by all HTTP/2 clients. Handlers should read before writing if
// possible to maximize compatibility.
func (aw *arrayWriter) Write(data []byte) (int, error) {
	// We could use a regexp to test for this, but the case we care about is
	// generagted by the gateway proxy so it shold be consistent.
	const itemsPrefix = `{"items":[`
	const itemsReplacement = "        ["

	// we lag one buffer behind to test for a gRPC array
	if aw.prevBuf == nil { // first write
		// if this looks like a gRPC array, blank out the initial key
		// to turn it into an anonymous array
		if bytes.HasPrefix(data, []byte(itemsPrefix)) {
			aw.prevBuf =
				bytes.Replace(
					data,
					[]byte(itemsPrefix),
					[]byte(itemsReplacement),
					1,
				)
			aw.isGrpcArray = true
		} else {
			aw.prevBuf = data
		}
		return 0, nil
	}

	n, err := aw.next.Write(aw.prevBuf)
	aw.prevBuf = data

	return n, err
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (aw *arrayWriter) WriteHeader(status int) {
	aw.status = status
	aw.next.WriteHeader(status)
}

// flush
func (aw *arrayWriter) flush() error {
	if (aw.status == 0 || aw.status == http.StatusOK) && aw.prevBuf != nil {
		var err error

		if aw.isGrpcArray {
			// don't write the final '}'
			_, err = aw.next.Write(aw.prevBuf[0 : len(aw.prevBuf)-1])

			// but do write something, so we don't change the content length
			if err == nil {
				_, err = aw.next.Write([]byte(" "))
			}

		} else {
			_, err = aw.next.Write(aw.prevBuf)
		}

		if err != nil {
			return errors.Wrap(err, "flush()")
		}
	}

	return nil
}
