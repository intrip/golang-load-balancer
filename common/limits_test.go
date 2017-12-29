package common

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

type FakeProxy struct{}

func (h *FakeProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintf(w, "hello")
}

func TestMaxConnections(t *testing.T) {
	maxConnections := 2
	serverUrl := "localhost:8088"

	go func() {
		s := &http.Server{
			Addr:    serverUrl,
			Handler: NewLimitHandler(maxConnections, &FakeProxy{}),
		}
		s.ListenAndServe()
	}()

	var wg sync.WaitGroup
	reachedMax := false
	wg.Add(maxConnections + 1)
	for i := 0; i <= maxConnections; i++ {
		go func() {
			defer wg.Done()
			res, _ := http.Get(fmt.Sprintf("http://%s/", serverUrl))
			if res.StatusCode == 503 {
				reachedMax = true
			}
		}()
	}
	wg.Wait()

	if !reachedMax {
		t.Errorf("Expected to reach maxConnections but did not")
	}
}
