package externalrefmock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

var (
	mu     sync.Mutex
	server *httptest.Server
)

type Response struct {
	PURL string `json:"purl"`
	URL  string `json:"url"`
	Kind string `json:"kind"`
}

func Start() *httptest.Server {
	mu.Lock()
	defer mu.Unlock()
	if server == nil {
		server = httptest.NewServer(http.HandlerFunc(handler))
	}
	return server
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if server != nil {
		server.Close()
		server = nil
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{
		PURL: r.URL.Query().Get("purl"),
		URL:  "https://github.com/example/repo",
		Kind: "vcs",
	})
}
