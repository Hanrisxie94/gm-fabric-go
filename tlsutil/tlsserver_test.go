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

package tlsutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func logNow(t *testing.T, format string, args ...interface{}) {
	// Tests can hang if t.Logf is used... so I use this
	t.Logf(format, args...)
}

func handleResult(t *testing.T, testResult chan bool, timeout chan bool, expected bool) {
	logNow(t, "* Test wait for a result")
	select {
	case result := <-testResult:
		{
			if result != expected {
				logNow(t, "! Test failed")
				t.FailNow()
			}
			logNow(t, "* Test got expected result")
		}
	case <-timeout:
		{
			logNow(t, "! Test timeout")
			t.FailNow()
		}
	}
}

func handleResultCreate() (chan bool, chan bool) {
	result := make(chan bool)
	timeout := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()
	return result, timeout
}

func TestHttps(t *testing.T) {
	testResult, timeout := handleResultCreate()
	rand.Seed(time.Now().Unix())
	port := int(6448 + rand.Int31n(100) + 100)

	logNow(t, "* Test create Server config....")

	serverTrust := "testcerts/server.trust.pem"
	serverCert := "testcerts/server.cert.pem"
	serverKey := "testcerts/server.key.pem"
	serverCN := "fabric-ssl-test-server"
	clientTrust := "testcerts/client.trust.pem"
	clientCert := "testcerts/client.cert.pem"
	clientKey := "testcerts/client.key.pem"

	cfg, err := BuildServerTLSConfig(
		serverTrust,
		serverCert,
		serverKey,
	)
	if err != nil {
		logNow(t, "! Test unable to get server config: %v", err)
		testResult <- false
		return
	}

	// Spawn an https handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cn := GetCommonName(GetDistinguishedName(r.TLS.PeerCertificates[0]))
		logNow(t, "* Identified client as %s", cn)
		w.WriteHeader(200)
		fmt.Fprintf(w, "ok ")
		fmt.Fprintf(w, cn)
	})
	host := "127.0.0.1"
	addr := fmt.Sprintf("%s:%d", host, port)
	httpServer := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: cfg,
	}
	logNow(t, "* Test https on %s", addr)
	go httpServer.ListenAndServeTLS(serverCert, serverKey)

	go func() {
		time.Sleep(2 * time.Second)
		logNow(t, "* Client spawning to talk to our server")
		url := fmt.Sprintf("https://%s/ping", addr)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logNow(t, "! Client unable to create request: %v", err)
			testResult <- false
			return
		}
		if req == nil {
			logNow(t, "! Client got nil request")
			testResult <- false
			return
		}
		// Perform an actual https connection with it
		clientConnFactory, err := NewTLSClientConnFactory(
			clientTrust,
			clientCert,
			clientKey,
			serverCN,
			host,
			fmt.Sprintf("%d", port),
		)
		if err != nil {
			logNow(t, "! Client unable to create connection factory: %v", err)
			testResult <- false
			return
		}
		if clientConnFactory == nil {
			logNow(t, "! Client conn factory is nil")
			testResult <- false
			return
		}
		res, err := clientConnFactory.Do(req)
		if err != nil {
			logNow(t, "! Client failed to connect: %v", err)
			testResult <- false
			return
		}
		if res.StatusCode != http.StatusOK {
			logNow(t, "! Client failed status: %d", res.StatusCode)
			testResult <- false
			return
		}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logNow(t, "! Client got bad response: %v", err)
			testResult <- false
			return
		}

		if string(bodyBytes[0:2]) != "ok" {
			logNow(t, "! Client got wrong response: %s", string(bodyBytes))
			testResult <- false
			return
		}
		logNow(t, "* Client Success for bodyBytes: %s", bodyBytes[2:])
		testResult <- true
		return
	}()

	handleResult(t, testResult, timeout, true)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	httpServer.Shutdown(ctx)
	cancelFunc()
}

func TestServerCorrect(t *testing.T) {
	// Seed the random number generator
	rand.Seed(time.Now().Unix())
	port := int(6448 + rand.Int31n(100))

	// Note:
	// * supply the CN of the server side if the server CN is not hostname - rather than InsecureSkipVerify
	// * There is no such thing as "use InsecureSkipVerify in test".
	//   It will make it into production if we don't code with InsecureSkipVerify off,
	//   because you need to plan to turn it off by supplying the serverCN as a parameter.
	// * trust is "symmectric"  client.trust.pem is things that signed server.cert.pem, vice versa.
	// * that we plugged in server.cert.pem as the *only* trusted cert.
	// We conclude the test by writing true or false to here
	testResult, timeout := handleResultCreate()

	serverWithConfig(
		t,
		port,
		testResult,
		"fabric-ssl-test-server",
		"testcerts/server.trust.pem", "testcerts/server.cert.pem", "testcerts/server.key.pem",
		"testcerts/client.trust.pem", "testcerts/client.cert.pem", "testcerts/client.key.pem",
	)

	handleResult(t, testResult, timeout, true)
}

func TestServerWrongServerCN(t *testing.T) {
	// Seed the random number generator
	rand.Seed(time.Now().Unix())
	port := int(6448 + rand.Int31n(100) + 100)
	testResult, timeout := handleResultCreate()

	serverWithConfig(
		t,
		port,
		testResult,
		"www.mitm.com",
		"testcerts/server.trust.pem", "testcerts/server.cert.pem", "testcerts/server.key.pem",
		"testcerts/client.trust.pem", "testcerts/client.cert.pem", "testcerts/client.key.pem",
	)

	handleResult(t, testResult, timeout, false)
}

func TestServerWrongCert(t *testing.T) {
	// Seed the random number generator
	rand.Seed(time.Now().Unix())
	port := int(6448 + rand.Int31n(100) + 200)
	testResult, timeout := handleResultCreate()

	serverWithConfig(
		t,
		port,
		testResult,
		"twl-server-generic2",
		"testcerts/server.trust.pem", "testcerts/wrong.cert.pem", "testcerts/wrong.key.pem",
		"testcerts/client.trust.pem", "testcerts/client.cert.pem", "testcerts/client.key.pem",
	)

	handleResult(t, testResult, timeout, false)
}

/*
  Test that the bidirectional TLS connection works.
  We do this concurrently, like real servers.
*/
func serverWithConfig(
	t *testing.T,
	port int,
	testResult chan<- bool,
	serverCN string,
	serverTrust, serverCert, serverKey string,
	clientTrust, clientCert, clientKey string,
) {

	logNow(t, "* Test create Server config....")
	cfg, err := BuildServerTLSConfig(serverTrust, serverCert, serverKey)
	if err != nil {
		logNow(t, "unable to get server config: %v", err)
		testResult <- false
		return
	}

	logNow(t, "* Test InsecureSkipVerify is %v", cfg.InsecureSkipVerify)

	host := "127.0.0.1"
	addr := fmt.Sprintf("%s:%d", host, port)
	logNow(t, "* Test create Client socket [%s]", addr)
	serverSock, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		logNow(t, "unable to listen on port: %v", err)
		testResult <- false
		return
	}

	go func() {
		logNow(t, "* Client send message")
		clientSock, err := NewTLSClientConn(
			clientTrust, clientCert, clientKey,
			serverCN,
			host,
			fmt.Sprintf("%d", port),
		)
		if err != nil {
			logNow(t, "! Client unable to connect: %v", err)
			testResult <- false
			return
		}

		go func() {
			logNow(t, "* Client sends a ping")
			_, err := clientSock.Write([]byte("ping"))
			if err != nil {
				logNow(t, "! Client failed to write to socket: %v", err)
				testResult <- false
				return
			}
			buf := make([]byte, 80)
			read, err := clientSock.Read(buf)
			if err != nil {
				logNow(t, "! Client failed to get a pong: %v", err)
				testResult <- false
				return
			}
			msg := string(buf[0:read])
			if msg != "pong" {
				logNow(t, "! Client didn't get pong")
				testResult <- false
				return
			}
			logNow(t, "* Client got a pong")
			testResult <- true
			return
		}()
	}()

	// This is a TLS server accepting connections and responding
	go func() {
		defer serverSock.Close()
		for {
			logNow(t, "* Server Accept")
			conn, err := serverSock.Accept()
			if err != nil {
				logNow(t, "! Server sock accept problem: %v", err)
				testResult <- false
				return
			}
			defer conn.Close()
			go func() {
				logNow(t, "* Server handling connection")
				buf := make([]byte, 80)
				read, err := conn.Read(buf)
				if err != nil {
					logNow(t, "! Server failed to get message from client: %v", err)
					testResult <- false
					return
				}
				msg := string(buf[0:read])
				if msg != "ping" {
					logNow(t, "! Server didn't get ping")
					testResult <- false
					return
				}
				logNow(t, "* Server got a ping")
				_, err = conn.Write([]byte("pong"))
				if err != nil {
					logNow(t, "! Server failed to send a pong: %v", err)
					testResult <- false
					return
				}
				logNow(t, "* Server sent a pong")
			}()
		}
	}()

}
