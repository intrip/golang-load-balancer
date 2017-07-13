package main

import (
	"flag"
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"io"
	"net"
	"os"
	"strings"
)

var backends []common.Backend

var (
	bind    = flag.String("bind", "0.0.0.0", "The address to bind on")
	port    = flag.Int("port", 8080, "The port to listen to")
	balance = flag.String("balancers", "0.0.0.0:8081", "The balancer as a csv list of ip:port")
)

var skipBalancing = false

func init() {
	flag.Parse()
	backends = parseBalance(*balance)
}

func main() {
	started := make(chan bool, 1)
	listen(bind, port, started)
}

func listen(bind *string, port *int, started chan bool) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *bind, *port))
	defer listener.Close()
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	started <- true

	for {
		conn, err := listener.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// needed for testing purpose
	if !skipBalancing {
		backendStrategy := &common.Backends{0, backends}
		next := backendStrategy.Next()
		doBalance(conn, &next)
	}
}

func parseBalance(balancers string) (backends []common.Backend) {
	backendsData := strings.Split(balancers, ",")
	backends = make([]common.Backend, len(backendsData))

	for index, backend := range backendsData {
		backendData := strings.SplitN(backend, ":", 2)
		backends[index] = common.Backend{backendData[0], backendData[1], 0}
	}

	return
}

func doBalance(fromConnection net.Conn, backend *common.Backend) {
	toConnection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", backend.Ip, backend.Port))
	if err != nil {
		fmt.Printf("Error connecting to %s:%s : %s\n", backend.Ip, backend.Port, err.Error())
		os.Exit(1)
	}
	defer toConnection.Close()

	backend.ActiveConnections++
	copy(fromConnection, toConnection)
}

func copy(from net.Conn, to net.Conn) {
	io.Copy(to, from)
}
