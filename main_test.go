package main

import (
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
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
	expectedBind := "0.0.0.0"
	expectedMaxConnections := 100
	expectedReadTimeout := 100
	expectedWriteTimeout := 100
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

func TestMaxConnections(t *testing.T) {

	//for i := 0; i <= maxConnections; i++ {
	//_, err := http.Get(fmt.Sprintf("http://%s/", serverUrl()))
	//if i == maxConnections && err == nil {
	//fmt.Println(err)
	//}
	//}
}

func TestDoBalance(t *testing.T) {
	msg := "Hello world!"

	// listen balancer
	beListen := "0.0.0.0:8081"
	beRemoteAddr := ""
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		beRemoteAddr = r.RemoteAddr
		doBalance(w, r, &common.Backend{Url: fmt.Sprintf("http://%s", beListen), ActiveConnections: 0})
	})

	// listen backend
	beServeMux := http.NewServeMux()
	beServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		expectedForwarded := fmt.Sprintf("by=%s; for=%s; host=%s; proto=%s", serverUrl(), beRemoteAddr, serverUrl(), r.Proto)
		if r.Header["Forwarded"][0] != expectedForwarded {
			t.Errorf("Expected to receive forwarded headers: %s, got: %s", r.Header["Forwarded"][0], expectedForwarded)
		}

		// send msg to the caller
		fmt.Fprintf(w, msg)
	})

	go func() {
		http.ListenAndServe(beListen, beServeMux)
	}()
	go func() {
		http.ListenAndServe(serverUrl(), serveMux)
	}()

	res, err := http.Get(fmt.Sprintf("http://%s/", serverUrl()))
	if err != nil {
		log.Panic("[test] Error connecting to balancer: ", err)
	}
	bodyBytes, _ := ioutil.ReadAll(res.Body)

	if string(bodyBytes) != msg {
		t.Errorf("Expected to read %s, got: %s", msg, bodyBytes)
	}
}
