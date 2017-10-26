package common

import (
	"net/http"
)

type limitHandler struct {
	connc   chan struct{}
	handler http.Handler
}

func (h *limitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	select {
	case <-h.connc:
		h.handler.ServeHTTP(w, req)
		h.connc <- struct{}{}
	default:
		http.Error(w, "503 too busy", 503)
	}
}

func NewLimitHandler(maxConns int, handler http.Handler) http.Handler {
	h := &limitHandler{
		connc:   make(chan struct{}, maxConns),
		handler: handler,
	}
	for i := 0; i < maxConns; i++ {
		h.connc <- struct{}{}
	}
	return h
}
