package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"
)

var DefaultAddress = "https://synchronization.werf.io"

func Run(_ context.Context, ip, port string, distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)) error {
	h := newHandler(distributedLockerBackendFactoryFunc)
	return http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), h)
}

type handler struct {
	*http.ServeMux

	distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)

	mux            sync.Mutex
	clientHandlers map[string]*clientHandler
}

func newHandler(distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)) *handler {
	srv := &handler{
		ServeMux:                            http.NewServeMux(),
		distributedLockerBackendFactoryFunc: distributedLockerBackendFactoryFunc,
		clientHandlers:                      make(map[string]*clientHandler),
	}
	srv.HandleFunc("/health", srv.handleHealth)
	srv.HandleFunc("/new-client-id", srv.handleNewClientID)
	srv.HandleFunc("/", srv.handleRequestByClientID)
	return srv
}

type healthRequest struct {
	Echo string `json:"echo"`
}

type healthResponse struct {
	Err    util.SerializableError `json:"err"`
	Echo   string                 `json:"echo"`
	Status string                 `json:"status"`
}

func (server *handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	var request healthRequest
	var response healthResponse

	handleRequest(w, r, &request, &response, func() {
		logboek.Debug().LogF("SynchronizationServerHandler -- Health request %#v\n", request)
		response.Echo = request.Echo
		response.Status = "OK"
		logboek.Debug().LogF("SynchronizationServerHandler -- Health response %#v\n", response)
	})
}

type (
	NewClientIDRequest  struct{}
	NewClientIDResponse struct {
		Err      util.SerializableError `json:"err"`
		ClientID string                 `json:"clientID"`
	}
)

func (server *handler) handleNewClientID(w http.ResponseWriter, r *http.Request) {
	var request NewClientIDRequest
	var response NewClientIDResponse
	handleRequest(w, r, &request, &response, func() {
		logboek.Debug().LogF("SynchronizationServerHandler -- NewClientID request %#v\n", request)
		response.ClientID = uuid.New().String()
		logboek.Debug().LogF("SynchronizationServerHandler -- NewClientID response %#v\n", response)
	})
}

func (server *handler) handleLanding(w http.ResponseWriter, r *http.Request) {
	rawPage := fmt.Sprintf(` <!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Werf synchronization http server</title>
</head>
<body>
<h1>Werf synchronization http server</h1>

<p>Werf uses --synchronization=%s as a default synchronization service.</p>

<p>Use "werf synchronization" command to run own synchronization http server. You can also configure werf to use local or kubernetes based synchronization backend.</p>
</body>
</html>
`, DefaultAddress)
	fmt.Fprintf(w, rawPage)
}

func (server *handler) handleRequestByClientID(w http.ResponseWriter, r *http.Request) {
	logboek.Debug().LogF("SynchronizationServerHandler -- ServeHTTP url path = %q\n", r.URL.Path)

	if r.URL.Path == "/" {
		server.handleLanding(w, r)
		return
	}

	clientID := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)[0]
	logboek.Debug().LogF("SynchronizationServerHandler -- ServeHTTP clientID = %q\n", clientID)

	if clientID == "" {
		http.Error(w, fmt.Sprintf("Bad request: cannot get clientID from URL path %q", r.URL.Path), http.StatusBadRequest)
		return
	}

	if clientServer, err := server.getOrCreateClientHandler(clientID); err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %s", err), http.StatusInternalServerError)
		return
	} else {
		http.StripPrefix(fmt.Sprintf("/%s", clientID), clientServer).ServeHTTP(w, r)
	}
}

func (server *handler) getOrCreateClientHandler(clientID string) (*clientHandler, error) {
	server.mux.Lock()
	defer server.mux.Unlock()

	if handler, hasKey := server.clientHandlers[clientID]; hasKey {
		return handler, nil
	} else {
		distributedLockerBackend, err := server.distributedLockerBackendFactoryFunc(clientID)
		if err != nil {
			return nil, fmt.Errorf("unable to create distributed locker backend for clientID %q: %w", clientID, err)
		}

		handler := newClientHandler(clientID, distributedLockerBackend)
		server.clientHandlers[clientID] = handler

		logboek.Debug().LogF("SynchronizationServerHandler -- Created new synchronization server handler by clientID %q: %v\n", clientID, handler)
		return handler, nil
	}
}

type clientHandler struct {
	*http.ServeMux
	ClientID string

	DistributedLockerBackend distributed_locker.DistributedLockerBackend
}

func newClientHandler(clientID string, distributedLockerBackend distributed_locker.DistributedLockerBackend) *clientHandler {
	srv := &clientHandler{
		ServeMux:                 http.NewServeMux(),
		ClientID:                 clientID,
		DistributedLockerBackend: distributedLockerBackend,
	}
	srv.Handle("/locker/", http.StripPrefix("/locker", distributed_locker.NewHttpBackendHandler(srv.DistributedLockerBackend)))
	return srv
}

func handleRequest(w http.ResponseWriter, r *http.Request, request, response interface{}, actionFunc func()) {
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, fmt.Sprintf("unable to unmarshal request json: %s", err), http.StatusBadRequest)
			return
		}
	}

	actionFunc()

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
