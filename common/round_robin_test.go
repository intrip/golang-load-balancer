package common

import (
	"testing"
)

func TestNext(t *testing.T) {
	backendA := Backend{Ip: "0.0.0.0", Port: "8081", ActiveConnections: 0}
	backendB := Backend{Ip: "0.0.0.0", Port: "8082", ActiveConnections: 0}
	backends := Backends{0, []Backend{backendA, backendB}}

	firstBackend := backends.Next()
	secondBackend := backends.Next()
	thirdBackend := backends.Next()

	if firstBackend != backendA {
		t.Errorf("Wrong order of Next(), expected %v, got: %v", backendA, firstBackend)
	}
	if secondBackend != backendB {
		t.Errorf("Wrong order of Next(), expected %v, got: %v", backendB, firstBackend)
	}
	if thirdBackend != backendA {
		t.Errorf("Wrong order of Next(), expected %v, got: %v", backendA, firstBackend)
	}
}
