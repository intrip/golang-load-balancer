package main

import (
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"io/ioutil"
	"log"
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
	expectedBackendsTimeout := 30
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
	if expectedBackendsTimeout != backendsTimeout {
		t.Errorf("BackendsTimeout differ, expected %d got %d", expectedBackendsTimeout, backendsTimeout)
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
		expectedForwarded := fmt.Sprintf("by=%s; for=%s; host=%s; proto=%s", serverUrl(), beRemoteAddr, serverUrl(), r.Proto)
		if r.Header["Forwarded"][0] != expectedForwarded {
			t.Errorf("Expected to receive forwarded headers: %s, got: %s", r.Header["Forwarded"][0], expectedForwarded)
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
		beRemoteAddr = r.RemoteAddr
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

// TODO consider using a mock instead
func TestBackendTimeout(t *testing.T) {
	oldBackendsTimeout := backendsTimeout
	backendsTimeout = 1
	// listen backend
	beListen := "localhost:8081"
	beRemoteAddr := ""
	beServeMux := http.NewServeMux()
	beServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(2) * time.Second)
	})
	beServer := &http.Server{
		Addr:    beListen,
		Handler: beServeMux,
	}

	// listen balancer
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		beRemoteAddr = r.RemoteAddr
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

	res, _ := http.Get(fmt.Sprintf("http://%s/", serverUrl()))

	if res.StatusCode != 503 {
		t.Errorf("Expected status code 503, got: %d", res.StatusCode)
	}

	defer beServer.Close()
	defer server.Close()
	backendsTimeout = oldBackendsTimeout
}

func TestBackendUnavailable(t *testing.T) {
	// listen balancer
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		doBalance(w, r, &common.Backend{Url: "http://0.0.0.0:9999", ActiveConnections: 0})
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

	if res.StatusCode != 503 {
		t.Errorf("Expected status code 503, got: %d", res.StatusCode)
	}

	defer server.Close()
}
