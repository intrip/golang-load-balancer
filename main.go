package main

import (
	"flag"
	"fmt"
	"github.com/intrip/simple_balancer/common"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	bind, balance        string
	port, maxConnections int
	readTimeout          = 10
	writeTimeout         = 10
	backends             []common.Backend
	testEnv              bool
)

func init() {
	loadConfig("config")

	if flag.Lookup("test.v") == nil {
		testEnv = false
	} else {
		testEnv = true
	}
}

// loads config from ./config.yaml
func loadConfig(config string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName(config)
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
		maxConnections, err = strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("Server maxConnections is not valid: %s \n", err))
		}
	} else {
		panic(fmt.Errorf("Server maxConnections is required"))
	}

	// timeout
	if v, ok := server["readtimeout"]; ok {
		readTimeout, err = strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("server readtimeout is not valid: %s \n", err))
		}
	}
	if v, ok := server["writetimeout"]; ok {
		writeTimeout, err = strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("server writetimeout is not valid: %s \n", err))
		}
	}
	balance = viper.GetString("balancers")
	backends = parseBalance(balance)
}

func main() {
	s := &http.Server{
		Addr:           serverUrl(),
		Handler:        common.NewLimitHandler(maxConnections, &Proxy{}),
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}

func serverUrl() string {
	return fmt.Sprintf("%s:%d", bind, port)
}

type Proxy struct{}

func (h *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backendStruct := &common.RoundRobin{0, backends}
	next := common.Next(backendStruct)
	doBalance(w, r, &next)
}

func doBalance(w http.ResponseWriter, r *http.Request, backend *common.Backend) {
	u, err := url.Parse(backend.Url)
	if err != nil {
		log.Panic("Error parsing backend Url: ", err)
	}

	if !testEnv {
		log.Printf("Request from: %s forwarded to: %s path: %s", r.RemoteAddr, backend.Url, r.RequestURI)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}

func parseBalance(balancers string) (backends []common.Backend) {
	urls := strings.Split(balancers, ",")
	backends = make([]common.Backend, len(urls))

	for index, backend := range urls {
		backends[index] = common.Backend{Url: backend, ActiveConnections: 0}
	}

	return
}
