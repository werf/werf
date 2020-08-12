package synchronization_server

import (
	"context"
	"net/http"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/storage"
)

func NewStagesStorageCacheHttpHandler(ctx context.Context, stagesStorageCache storage.StagesStorageCache) *StagesStorageCacheHttpHandler {
	handler := &StagesStorageCacheHttpHandler{
		StagesStorageCache: stagesStorageCache,
		ServeMux:           http.NewServeMux(),
	}
	handler.HandleFunc("/get-all-stages", handler.handleGetAllStages(ctx))
	handler.HandleFunc("/delete-all-stages", handler.handleDeleteAllStages(ctx))
	handler.HandleFunc("/get-stages-by-signature", handler.handleGetStagesBySignature(ctx))
	handler.HandleFunc("/store-stages-by-signature", handler.handleStoreStagesBySignature(ctx))
	handler.HandleFunc("/delete-stages-by-signature", handler.handleDeleteStagesBySignature(ctx))

	return handler
}

type StagesStorageCacheHttpHandler struct {
	*http.ServeMux
	StagesStorageCache storage.StagesStorageCache
}

type GetAllStagesRequest struct {
	ProjectName string `json:"projectName"`
}
type GetAllStagesResponse struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []image.StageID        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandler) handleGetAllStages(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetAllStagesRequest
		var response GetAllStagesResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- GetAllStages request %#v\n", request)
			response.Found, response.Stages, response.Err.Error = handler.StagesStorageCache.GetAllStages(ctx, request.ProjectName)
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- GetAllStages response %#v\n", response)
		})
	}
}

type DeleteAllStagesRequest struct {
	ProjectName string `json:"projectName"`
}
type DeleteAllStagesResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleDeleteAllStages(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteAllStagesRequest
		var response DeleteAllStagesResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- DeleteAllStages request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteAllStages(ctx, request.ProjectName)
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- DeleteAllStages response %#v\n", response)
		})
	}
}

type GetStagesBySignatureRequest struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}
type GetStagesBySignatureResponse struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []image.StageID        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandler) handleGetStagesBySignature(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetStagesBySignatureRequest
		var response GetStagesBySignatureResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- GetStagesBySignature request %#v\n", request)
			response.Found, response.Stages, response.Err.Error = handler.StagesStorageCache.GetStagesBySignature(ctx, request.ProjectName, request.Signature)
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- GetStagesBySignature response %#v\n", response)
		})
	}
}

type StoreStagesBySignatureRequest struct {
	ProjectName string          `json:"projectName"`
	Signature   string          `json:"signature"`
	Stages      []image.StageID `json:"stages"`
}
type StoreStagesBySignatureResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleStoreStagesBySignature(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request StoreStagesBySignatureRequest
		var response StoreStagesBySignatureResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- StoreStagesBySignature request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.StoreStagesBySignature(ctx, request.ProjectName, request.Signature, request.Stages)
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- StoreStagesBySignature response %#v\n", response)
		})
	}
}

type DeleteStagesBySignatureRequest struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}
type DeleteStagesBySignatureResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandler) handleDeleteStagesBySignature(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteStagesBySignatureRequest
		var response DeleteStagesBySignatureResponse
		HandleRequest(w, r, &request, &response, func() {
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- DeleteStagesBySignature request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteStagesBySignature(ctx, request.ProjectName, request.Signature)
			logboek.Context(ctx).Debug().LogF("StagesStorageCacheHttpHandler -- DeleteStagesBySignature response %#v\n", response)
		})
	}
}
