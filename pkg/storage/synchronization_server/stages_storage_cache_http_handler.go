package synchronization_server

import (
	"context"
	"net/http"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

func NewStagesStorageCacheHttpHandler(stagesStorageCache StagesStorageCacheInterface) *StagesStorageCacheHttpHandler {
	handler := &StagesStorageCacheHttpHandler{
		StagesStorageCache: stagesStorageCache,
		ServeMux:           http.NewServeMux(),
	}
	handler.HandleFunc("/get-all-stages", handler.handleGetAllStages())
	handler.HandleFunc("/delete-all-stages", handler.handleDeleteAllStages())
	handler.HandleFunc("/get-stages-by-digest", handler.handleGetStagesByDigest())
	handler.HandleFunc("/store-stages-by-digest", handler.handleStoreStagesByDigest())
	handler.HandleFunc("/delete-stages-by-digest", handler.handleDeleteStagesByDigest())
	return handler
}

type StagesStorageCacheHttpHandler struct {
	*http.ServeMux
	StagesStorageCache StagesStorageCacheInterface
}

type GetAllStagesRequest struct {
	ProjectName string `json:"projectName"`
}

type GetAllStagesResponse struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []image.StageID        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandler) handleGetAllStages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetAllStagesRequest
		var response GetAllStagesResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- GetAllStages request %#v\n", request)
			response.Found, response.Stages, response.Err.Error = handler.StagesStorageCache.GetAllStages(context.Background(), request.ProjectName)
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- GetAllStages response %#v\n", response)
		})
	}
}

type DeleteAllStagesRequest struct {
	ProjectName string `json:"projectName"`
}

type DeleteAllStagesResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleDeleteAllStages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteAllStagesRequest
		var response DeleteAllStagesResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- DeleteAllStages request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteAllStages(context.Background(), request.ProjectName)
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- DeleteAllStages response %#v\n", response)
		})
	}
}

type GetStagesByDigestRequest struct {
	ProjectName string `json:"projectName"`
	Digest      string `json:"digest"`
}

type GetStagesByDigestResponse struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []image.StageID        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandler) handleGetStagesByDigest() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetStagesByDigestRequest
		var response GetStagesByDigestResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- GetStagesByDigest request %#v\n", request)
			response.Found, response.Stages, response.Err.Error = handler.StagesStorageCache.GetStagesByDigest(context.Background(), request.ProjectName, request.Digest)
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- GetStagesByDigest response %#v\n", response)
		})
	}
}

type StoreStagesByDigestRequest struct {
	ProjectName string          `json:"projectName"`
	Digest      string          `json:"digest"`
	Stages      []image.StageID `json:"stages"`
}

type StoreStagesByDigestResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleStoreStagesByDigest() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request StoreStagesByDigestRequest
		var response StoreStagesByDigestResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- StoreStagesByDigest request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.StoreStagesByDigest(context.Background(), request.ProjectName, request.Digest, request.Stages)
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- StoreStagesByDigest response %#v\n", response)
		})
	}
}

type DeleteStagesByDigestRequest struct {
	ProjectName string `json:"projectName"`
	Digest      string `json:"digest"`
}

type DeleteStagesByDigestResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleDeleteStagesByDigest() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteStagesByDigestRequest
		var response DeleteStagesByDigestResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- DeleteStagesByDigest request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteStagesByDigest(context.Background(), request.ProjectName, request.Digest)
			logboek.Debug().LogF("StagesStorageCacheHttpHandler -- DeleteStagesByDigest response %#v\n", response)
		})
	}
}
