package synchronization_server

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/storage"
)

func RunSynchronizationServer(ip, port string, lockManagerFactoryFunc func(clientID string) (storage.LockManager, error), stagesStorageCacheFactoryFunc func(clientID string) (storage.StagesStorageCache, error)) error {
	handler := NewSynchronizationServerHandler(lockManagerFactoryFunc, stagesStorageCacheFactoryFunc)
	return http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), handler)
}

type SynchronizationServerHandler struct {
	LockManagerFactoryFunc        func(clientID string) (storage.LockManager, error)
	StagesStorageCacheFactoryFunc func(clientID string) (storage.StagesStorageCache, error)

	mux                             sync.Mutex
	SyncrhonizationServerByClientID map[string]*SynchronizationServerHandlerByClientID
}

func NewSynchronizationServerHandler(lockManagerFactoryFunc func(requestID string) (storage.LockManager, error), stagesStorageCacheFactoryFunc func(requestID string) (storage.StagesStorageCache, error)) *SynchronizationServerHandler {
	return &SynchronizationServerHandler{
		LockManagerFactoryFunc:          lockManagerFactoryFunc,
		StagesStorageCacheFactoryFunc:   stagesStorageCacheFactoryFunc,
		SyncrhonizationServerByClientID: make(map[string]*SynchronizationServerHandlerByClientID),
	}
}

func (server *SynchronizationServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logboek.Debug.LogF("SynchronizationServerHandler -- ServeHTTP url path = %q\n", r.URL.Path)

	clientID := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)[0]
	logboek.Debug.LogF("SynchronizationServerHandler -- ServeHTTP clientID = %q\n", clientID)

	if clientID == "" {
		http.Error(w, fmt.Sprintf("Bad request: cannot get clientID from URL path %q", r.URL.Path), http.StatusBadRequest)
		return
	}

	if clientServer, err := server.getOrCreateHandlerByClientID(clientID); err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %s", err), http.StatusInternalServerError)
		return
	} else {
		fmt.Printf("clientServer=%#v\n", clientServer)
		http.StripPrefix(fmt.Sprintf("/%s", clientID), clientServer).ServeHTTP(w, r)
	}
}

func (server *SynchronizationServerHandler) getOrCreateHandlerByClientID(clientID string) (*SynchronizationServerHandlerByClientID, error) {
	server.mux.Lock()
	defer server.mux.Unlock()

	if handler, hasKey := server.SyncrhonizationServerByClientID[clientID]; hasKey {
		return handler, nil
	} else {
		lockManager, err := server.LockManagerFactoryFunc(clientID)
		if err != nil {
			return nil, fmt.Errorf("unable to create lock manager for clientID %q: %s", clientID, err)
		}

		stagesStorageCache, err := server.StagesStorageCacheFactoryFunc(clientID)
		if err != nil {
			return nil, fmt.Errorf("unable to create stages storage cache for clientID %q: %s", clientID, err)
		}

		handler := NewSynchronizationServerHandlerByClientID(clientID, lockManager, stagesStorageCache)
		server.SyncrhonizationServerByClientID[clientID] = handler

		logboek.Debug.LogF("SynchronizationServerHandler -- Created new syncrhonization server handler by clientID %q: %v\n", clientID, handler)
		return handler, nil
	}
}

type SynchronizationServerHandlerByClientID struct {
	*http.ServeMux
	ClientID           string
	LockManager        storage.LockManager
	StagesStorageCache storage.StagesStorageCache
}

func NewSynchronizationServerHandlerByClientID(clientID string, lockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *SynchronizationServerHandlerByClientID {
	srv := &SynchronizationServerHandlerByClientID{
		ServeMux:           http.NewServeMux(),
		ClientID:           clientID,
		LockManager:        lockManager,
		StagesStorageCache: stagesStorageCache,
	}

	srv.Handle("/lock-manager/", http.StripPrefix("/lock-manager", NewLockManagerHttpHandler(lockManager)))
	srv.Handle("/stages-storage-cache/", http.StripPrefix("/stages-storage-cache", NewStagesStorageCacheHttpHandler(stagesStorageCache)))

	return srv
}
