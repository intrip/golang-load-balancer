package main

import (
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"github.com/spf13/viper"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
)

var (
	bind, balance                     string
	port                              int
	maxConnections, activeConnections int32
	skipBalancing                     = false
	backends                          []common.Backend
	done                              = false
)

const (
	MAX_CONN = "Max connection reached, closing connection.\n"
)

func init() {
	loadConfig()
	activeConnections = 0
}

// loads config from ./config.yaml
func loadConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error in config file: %s \n", err))
	}

	server := viper.GetStringMapString("server")
	// read port
	if v, ok := server["port"]; ok {
		port, err = strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("Server port is not valid: %s \n", err))
		}
	} else {
		panic(fmt.Errorf("Server port is required"))
	}
	// listen
	if v, ok := server["bind"]; ok {
		bind = v
	} else {
		panic(fmt.Errorf("Server bind is required"))
	}
	// maxConnections
	if v, ok := server["maxconnections"]; ok {
		maxConnInt, err := strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("Server maxConnections is not valid: %s \n", err))
		}
		maxConnections = int32(maxConnInt)

	} else {
		panic(fmt.Errorf("Server maxConnections is required"))
	}

	balance = viper.GetString("balancers")
	backends = parseBalance(balance)
}

func main() {
	// here we don't need to wait for started as we do in the tests
	started := make(chan bool, 1)
	listen(bind, port, started)
}

func listen(bind string, port int, started chan bool) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bind, port))
	if err != nil {
		panic(fmt.Errorf("Error listening:", err.Error()))
	}
	defer listener.Close()
	// used for testing purpose
	started <- true

	backendStruct := &common.Backends{0, backends}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		atomic.AddInt32(&activeConnections, 1)
		if activeConnections > maxConnections {
			conn.Close()
		} else {
			go handleConnection(conn, backendStruct)
		}
	}
}

func handleConnection(conn net.Conn, backendStruct *common.Backends) {
	defer conn.Close()
	defer func() { atomic.AddInt32(&activeConnections, -1) }()

	// needed for testing purpose
	if !skipBalancing {
		next := common.NextRoundRobin(backendStruct)
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
		panic(fmt.Errorf("Error connecting to %s:%s : %s\n", backend.Ip, backend.Port, err.Error()))
	}
	defer toConnection.Close()

	backend.ActiveConnections++
	copy(fromConnection, toConnection)
}

func copy(from net.Conn, to net.Conn) {
	io.Copy(to, from)
}
