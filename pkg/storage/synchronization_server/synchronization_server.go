package synchronization_server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
)

func RunSynchronizationServer(_ context.Context, ip, port string, distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error), stagesStorageCacheFactoryFunc func(clientID string) (StagesStorageCacheInterface, error)) error {
	handler := NewSynchronizationServerHandler(distributedLockerBackendFactoryFunc, stagesStorageCacheFactoryFunc)
	return http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), handler)
}

type SynchronizationServerHandler struct {
	*http.ServeMux

	DistributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)
	StagesStorageCacheFactoryFunc       func(clientID string) (StagesStorageCacheInterface, error)

	mux                             sync.Mutex
	SynchronizationServerByClientID map[string]*SynchronizationServerHandlerByClientID
}

func NewSynchronizationServerHandler(distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error), stagesStorageCacheFactoryFunc func(requestID string) (StagesStorageCacheInterface, error)) *SynchronizationServerHandler {
	srv := &SynchronizationServerHandler{
		ServeMux:                            http.NewServeMux(),
		DistributedLockerBackendFactoryFunc: distributedLockerBackendFactoryFunc,
		StagesStorageCacheFactoryFunc:       stagesStorageCacheFactoryFunc,
		SynchronizationServerByClientID:     make(map[string]*SynchronizationServerHandlerByClientID),
	}
	srv.HandleFunc("/health", srv.handleHealth)
	srv.HandleFunc("/new-client-id", srv.handleNewClientID)
	srv.HandleFunc("/", srv.handleRequestByClientID)
	return srv
}

type HealthRequest struct {
	Echo string `json:"echo"`
}

type HealthResponse struct {
	Err    util.SerializableError `json:"err"`
	Echo   string                 `json:"echo"`
	Status string                 `json:"status"`
}

func (server *SynchronizationServerHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	var request HealthRequest
	var response HealthResponse

	HandleRequest(w, r, &request, &response, func() {
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

func (server *SynchronizationServerHandler) handleNewClientID(w http.ResponseWriter, r *http.Request) {
	var request NewClientIDRequest
	var response NewClientIDResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug().LogF("SynchronizationServerHandler -- NewClientID request %#v\n", request)
		response.ClientID = uuid.New().String()
		logboek.Debug().LogF("SynchronizationServerHandler -- NewClientID response %#v\n", response)
	})
}

func (server *SynchronizationServerHandler) handleLanding(w http.ResponseWriter, r *http.Request) {
	rawPage := ` <!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Werf synchronization http server</title>
</head>
<body>
<h1>Werf synchronization http server</h1>

<p>Werf uses --synchronization=https://synchronization.werf.io as a default synchronization service.</p>

<p>Use "werf synchronization" command to run own synchronization http server. You can also configure werf to use local or kubernetes based synchronization backend.</p>

<p>More info about synchronization in werf: <a href="https://werf.io/documentation/internals/stages_and_storage.html#synchronization-locks-and-stages-storage-cache">https://werf.io/documentation/internals/stages_and_storage.html#synchronization-locks-and-stages-storage-cache</a></p>
</body>
</html>
`
	fmt.Fprintf(w, rawPage)
}

func (server *SynchronizationServerHandler) handleRequestByClientID(w http.ResponseWriter, r *http.Request) {
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

	if clientServer, err := server.getOrCreateHandlerByClientID(clientID); err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %s", err), http.StatusInternalServerError)
		return
	} else {
		http.StripPrefix(fmt.Sprintf("/%s", clientID), clientServer).ServeHTTP(w, r)
	}
}

func (server *SynchronizationServerHandler) getOrCreateHandlerByClientID(clientID string) (*SynchronizationServerHandlerByClientID, error) {
	server.mux.Lock()
	defer server.mux.Unlock()

	if handler, hasKey := server.SynchronizationServerByClientID[clientID]; hasKey {
		return handler, nil
	} else {
		distributedLockerBackend, err := server.DistributedLockerBackendFactoryFunc(clientID)
		if err != nil {
			return nil, fmt.Errorf("unable to create distributed locker backend for clientID %q: %w", clientID, err)
		}

		stagesStorageCache, err := server.StagesStorageCacheFactoryFunc(clientID)
		if err != nil {
			return nil, fmt.Errorf("unable to create stages storage cache for clientID %q: %w", clientID, err)
		}

		handler := NewSynchronizationServerHandlerByClientID(clientID, distributedLockerBackend, stagesStorageCache)
		server.SynchronizationServerByClientID[clientID] = handler

		logboek.Debug().LogF("SynchronizationServerHandler -- Created new synchronization server handler by clientID %q: %v\n", clientID, handler)
		return handler, nil
	}
}

type SynchronizationServerHandlerByClientID struct {
	*http.ServeMux
	ClientID string

	DistributedLockerBackend distributed_locker.DistributedLockerBackend
	StagesStorageCache       StagesStorageCacheInterface
}

func NewSynchronizationServerHandlerByClientID(clientID string, distributedLockerBackend distributed_locker.DistributedLockerBackend, stagesStorageCache StagesStorageCacheInterface) *SynchronizationServerHandlerByClientID {
	srv := &SynchronizationServerHandlerByClientID{
		ServeMux:                 http.NewServeMux(),
		ClientID:                 clientID,
		DistributedLockerBackend: distributedLockerBackend,
		StagesStorageCache:       stagesStorageCache,
	}
	srv.Handle("/locker/", http.StripPrefix("/locker", distributed_locker.NewHttpBackendHandler(srv.DistributedLockerBackend)))
	srv.Handle("/stages-storage-cache/v1/", http.StripPrefix("/stages-storage-cache/v1", NewStagesStorageCacheHttpHandler(stagesStorageCache)))
	srv.Handle("/stages-storage-cache/", http.StripPrefix("/stages-storage-cache", NewStagesStorageCacheHttpHandlerLegacy(stagesStorageCache)))
	return srv
}
