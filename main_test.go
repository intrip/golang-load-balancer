package main

import (
	"fmt"
	"github.com/intrip/golang-load-balancer/common"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestParseBalance(t *testing.T) {
	balance := "http://0.0.0.0:3000/home,http://0.0.0.0:3001/info"
	expectedBackends := []common.Backend{common.Backend{"http://0.0.0.0:3000/home", 0}, common.Backend{"http://0.0.0.0:3001/info", 0}}

	backends := parseBalance(balance)

	if len(backends) != len(expectedBackends) {
		t.Errorf("Result size differ, expected: %d, got: %d", len(expectedBackends), len(backends))
		return
	}

	for index, backend := range backends {
		if backend != expectedBackends[index] {
			t.Errorf("Backend %d differ, expected: %q got %q", index, expectedBackends[index], backend)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	loadConfig("config_test")
	expectedPort := 8080
	expectedBind := "localhost"
	expectedMaxConnections := 100
	expectedReadTimeout := 30
	expectedWriteTimeout := 30
	expectedBalance := "http://0.0.0.0:8081"

	if expectedPort != port {
		t.Errorf("Port differ, expected %d got %d", expectedPort, port)
	}
	if expectedBind != bind {
		t.Errorf("Bind differ, expected %d got %d", expectedBind, bind)
	}
	if expectedMaxConnections != maxConnections {
		t.Errorf("MaxConnections differ, expected %d got %d", expectedMaxConnections, maxConnections)
	}
	if expectedReadTimeout != readTimeout {
		t.Errorf("readTimeout differ, expected %d got %d", expectedReadTimeout, readTimeout)
	}
	if expectedWriteTimeout != writeTimeout {
		t.Errorf("WriteTimeout differ, expected %d got %d", expectedWriteTimeout, writeTimeout)
	}
	if expectedBalance != balance {
		t.Errorf("Balance differ, expected %d got %d", expectedBalance, balance)
	}
}

func TestDoBalance(t *testing.T) {
	msg := "Hello world!"

	// listen backend
	beListen := "localhost:8081"
	beRemoteAddr := ""
	beServeMux := http.NewServeMux()
	beServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header["X-Forwarded-For"][0] != beRemoteAddr {
			t.Errorf("Expected X-Forwarded-For: %s, got: %s", r.Header["X-Forwarded-For"][0], beRemoteAddr)
		}

		// send msg to the caller
		fmt.Fprintf(w, msg)
	})
	beServer := &http.Server{
		Addr:    beListen,
		Handler: beServeMux,
	}

	// listen balancer
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		beRemoteAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
		doBalance(w, r, &common.Backend{Url: fmt.Sprintf("http://%s", beListen), ActiveConnections: 0})
	})
	server := &http.Server{
		Addr:    serverUrl(),
		Handler: serveMux,
	}

	// backend
	go func() {
		beServer.ListenAndServe()
	}()
	// balancer
	go func() {
		server.ListenAndServe()
	}()

	time.Sleep(time.Duration(100) * time.Millisecond)
	res, err := http.Get(fmt.Sprintf("http://%s/", serverUrl()))
	if err != nil {
		log.Panic("[test] Error connecting to balancer: ", err)
	}
	bodyBytes, _ := ioutil.ReadAll(res.Body)

	if string(bodyBytes) != msg {
		t.Errorf("Expected to read %s, got: %s", msg, bodyBytes)
	}

	defer beServer.Close()
	defer server.Close()
}

func TestBackendUnavailable(t *testing.T) {
	// backend
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		doBalance(w, r, &common.Backend{Url: "http://0.0.0.0:9999"})
	})
	server := &http.Server{
		Addr:    serverUrl(),
		Handler: serveMux,
	}
	// balancer
	go func() {
		server.ListenAndServe()
	}()

	res, _ := http.Get(fmt.Sprintf("http://%s/", serverUrl()))

	if res.StatusCode != 502 {
		t.Errorf("Expected status code 502, got: %d", res.StatusCode)
	}

	defer server.Close()
}
