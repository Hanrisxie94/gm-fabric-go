package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type handlerConfig struct {
	maxBodySize int
	minDuration time.Duration
	maxDuration time.Duration
}

func main() {

	hiddenPort := getEnv("PMTEST_HIDDEN_PORT", "3000")
	clientCount, err := strconv.Atoi(getEnv("PMTEST_CLIENT_COUNT", "2"))
	if err != nil {
		log.Fatalf("Invalid PMTEST_CLIENT_COUNT '%s'", getEnv("PMTEST_CLIENT_COUNT", "2"))
	}
	maxBodySize, err := strconv.Atoi(getEnv("PMTEST_MAX_BODY_SIZE", "1048576"))
	if err != nil {
		log.Fatalf("Invalid PMTEST_MAX_BODY_SIZE '%s'", getEnv("PMTEST_MAX_BODY_SIZE", "1048576"))
	}
	exposedAddress := getEnv("PMTEST_EXPOSED_ADDRESS", "127.0.0.1:3000")
	testURLs := []string{
		fmt.Sprintf("http://%s/slow", exposedAddress),
		fmt.Sprintf("http://%s/medium", exposedAddress),
		fmt.Sprintf("http://%s/fast", exposedAddress),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/slow",
		handlerFactory(
			handlerConfig{
				maxBodySize: maxBodySize,
				minDuration: 5 * time.Second,
				maxDuration: 10 * time.Second,
			},
		),
	)
	mux.HandleFunc(
		"/medium",
		handlerFactory(
			handlerConfig{
				maxBodySize: maxBodySize,
				minDuration: 2 * time.Second,
				maxDuration: 5 * time.Second,
			},
		),
	)
	mux.HandleFunc(
		"/fast",
		handlerFactory(
			handlerConfig{
				maxBodySize: maxBodySize,
				minDuration: 0 * time.Second,
				maxDuration: 2 * time.Second,
			},
		),
	)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", hiddenPort),
		Handler: mux,
	}

	log.Printf("server listening on :%s", hiddenPort)
	go func() {
		log.Printf("server terminates: %s", server.ListenAndServe())
	}()

	log.Printf("starting %d test clients", clientCount)

	haltChan := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		clientID := i + 1
		go func() {
			defer wg.Done()
			testClient(clientID, testURLs, maxBodySize, haltChan)
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s := <-sigChan
	log.Printf("program received %v: terminating", s)

	log.Printf("waiting for clients to terminate")
	close(haltChan)
	wg.Wait()

	log.Printf("waiting for server to terminate")
	server.Close()

	log.Printf("done")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func handlerFactory(cfg handlerConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		handler(cfg, w, req)
	}
}

func handler(
	cfg handlerConfig,
	w http.ResponseWriter,
	req *http.Request,
) {
	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("ERROR: reading body:  %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("read %d bytes from body", len(body))
		if err := req.Body.Close(); err != nil {
			log.Printf("ERROR: req,Body.Close failed: %s", err)
		}
	}

	delayInterval := cfg.maxDuration - cfg.minDuration
	delayIncrement := rand.Int63n(int64(delayInterval))
	delay := cfg.minDuration + time.Duration(delayIncrement)

	bodySize := rand.Intn(cfg.maxBodySize)

	log.Printf("delay %s, return body size = %d", delay, bodySize)
	time.Sleep(delay)

	if _, err := w.Write(bytes.Repeat([]byte{'*'}, bodySize)); err != nil {
		log.Printf("ERROR: writing body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func testClient(
	clientID int,
	testURLs []string,
	maxBodySize int,
	haltChan <-chan struct{},
) {
CLIENT_LOOP:
	for tr := 1; ; tr++ {

		select {
		case <-haltChan:
			log.Printf("client %2d: tr %5d: haltChan closed", clientID, tr)
			break CLIENT_LOOP
		default:
		}

		url := testURLs[rand.Int()%len(testURLs)]
		bodySize := rand.Intn(maxBodySize)

		log.Printf("client %2d: tr %5d: requesting URL %s, body size = %d",
			clientID, tr, url, bodySize,
		)

		var client http.Client

		body := bytes.Repeat([]byte{'*'}, bodySize)
		req, err := http.NewRequest("GET", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("client %2d: tr %5d: ERROR: http.NewRequest failed: %s",
				clientID, tr, err,
			)
			break CLIENT_LOOP
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("client %2d: tr %5d: ERROR: client.Get failed: %s",
				clientID, tr, err,
			)
			break CLIENT_LOOP
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("client %2d: tr %5d: ERROR: unexpected status: (%d) %s",
				clientID, tr, resp.StatusCode, resp.Status,
			)
			break CLIENT_LOOP
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("client %2d: tr %5d: ERROR: ioutil.ReadAll(resp.Body) failed: %s",
				clientID, tr, err,
			)
			break CLIENT_LOOP
		}
		if err = resp.Body.Close(); err != nil {
			log.Printf("client %2d: tr %5d: ERROR: resp.Body.Close() failed: %s",
				clientID, tr, err,
			)
			break CLIENT_LOOP
		}

		log.Printf("client %2d: tr %5d: body size = %d", clientID, tr, len(body))
	}

}
