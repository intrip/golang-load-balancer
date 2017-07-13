package common

type Backend struct {
	Ip                string
	Port              string
	ActiveConnections int
}

type Backends struct {
	LastSelected Backend
	Backends     []Backend
}

