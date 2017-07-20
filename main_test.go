package main

import (
	"bufio"
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"net"
	"strconv"
	"testing"
)

func TestAcceptsLocalConnections(t *testing.T) {
	// here we don't buffer the channel to wait until has started
	startedListening := make(chan bool)
	skipBalancing = true
	go listen(bind, port, startedListening)
	<-startedListening

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", bind, port))
	defer conn.Close()
	if err != nil {
		t.Errorf("the server does not accept local connections: %s", err.Error())
	}
}

func TestParseBalance(t *testing.T) {
	balance := "0.0.0.0:3000,0.0.0.0:3001"
	expectedBackends := [2]common.Backend{common.Backend{"0.0.0.0", "3000", 0}, common.Backend{"0.0.0.0", "3001", 0}}

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

func TestDoBalance(t *testing.T) {
	// balancer info
	balancerIp := "0.0.0.0"
	balancerPort := 3000
	// backend info
	beIp := "0.0.0.0"
	bePort := 3001
	toBackend := common.Backend{beIp, strconv.Itoa(bePort), 0}
	testMessage := "Hello world!"

	done := make(chan bool)
	connection := make(chan net.Conn)
	// start listening balancer
	go func(done chan bool, connection chan net.Conn) {
		listener, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", balancerIp, balancerPort))
		defer listener.Close()
		done <- true
		conn, _ := listener.Accept()
		connection <- conn
	}(done, connection)
	// start listening backend
	go func(done chan bool, connection chan net.Conn) {
		listener, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", beIp, bePort))
		defer listener.Close()
		done <- true
		conn, _ := listener.Accept()
		connection <- conn
	}(done, connection)

	// wait for both listeners
	<-done
	<-done

	// connects to the balancer and send test data
	balancerConnTo, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", balancerIp, balancerPort))
	writer := bufio.NewWriter(balancerConnTo)
	writer.WriteString(fmt.Sprintf("%s\n", testMessage))
	writer.Flush()

	balancerConnFrom := <-connection
	go doBalance(balancerConnFrom, &toBackend)

	backendConn := <-connection
	reader := bufio.NewReader(backendConn)
	msg, _ := reader.ReadString('\n')
	if msg[:len(msg)-1] != testMessage {
		t.Errorf("Message received is wrong, expected: %s got: %s", testMessage, msg)
	}

	// verify that increments active connections
	if toBackend.ActiveConnections == 0 {
		t.Errorf("Expected activeConnection to be 1, got: %d", toBackend.ActiveConnections)
	}
}
