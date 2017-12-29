package common

import (
	"testing"
)

func TestNext(t *testing.T) {
	backendA := Backend{Url: "http://localhost:8081", ActiveConnections: 0}
	backendB := Backend{Url: "http://localhost:8082", ActiveConnections: 0}
	backends := RoundRobin{0, []Backend{backendA, backendB}}

	firstBackend := Next(&backends)
	secondBackend := Next(&backends)
	thirdBackend := Next(&backends)

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
