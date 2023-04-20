package synchronization_server

import (
	"context"
	"net/http"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type StageIDLegacy struct {
	Signature string `json:"signature"`
	UniqueID  int64  `json:"uniqueID"`
}

func NewStagesStorageCacheHttpHandlerLegacy(stagesStorageCache StagesStorageCacheInterface) *StagesStorageCacheHttpHandlerLegacy {
	handler := &StagesStorageCacheHttpHandlerLegacy{
		StagesStorageCache: stagesStorageCache,
		ServeMux:           http.NewServeMux(),
	}
	handler.HandleFunc("/get-all-stages", handler.handleGetAllStages())
	handler.HandleFunc("/delete-all-stages", handler.handleDeleteAllStages())
	handler.HandleFunc("/get-stages-by-signature", handler.handleGetStagesBySignature())
	handler.HandleFunc("/store-stages-by-signature", handler.handleStoreStagesBySignature())
	handler.HandleFunc("/delete-stages-by-signature", handler.handleDeleteStagesBySignature())
	return handler
}

type StagesStorageCacheHttpHandlerLegacy struct {
	*http.ServeMux
	StagesStorageCache StagesStorageCacheInterface
}

type GetAllStagesRequestLegacy struct {
	ProjectName string `json:"projectName"`
}

type GetAllStagesResponseLegacy struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []StageIDLegacy        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandlerLegacy) handleGetAllStages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetAllStagesRequestLegacy
		var response GetAllStagesResponseLegacy
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- GetAllStages request %#v\n", request)
			found, stages, err := handler.StagesStorageCache.GetAllStages(context.Background(), request.ProjectName)

			for _, s := range stages {
				response.Stages = append(response.Stages, StageIDLegacy{Signature: s.Digest, UniqueID: s.UniqueID})
			}
			response.Found = found
			response.Err.Error = err

			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- GetAllStages response %#v\n", response)
		})
	}
}

type DeleteAllStagesRequestLegacy struct {
	ProjectName string `json:"projectName"`
}

type DeleteAllStagesResponseLegacy struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandlerLegacy) handleDeleteAllStages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteAllStagesRequestLegacy
		var response DeleteAllStagesResponseLegacy
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- DeleteAllStages request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteAllStages(context.Background(), request.ProjectName)
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- DeleteAllStages response %#v\n", response)
		})
	}
}

type GetStagesBySignatureRequestLegacy struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}

type GetStagesBySignatureResponseLegacy struct {
	Err    util.SerializableError `json:"err"`
	Found  bool                   `json:"found"`
	Stages []StageIDLegacy        `json:"stages"`
}

func (handler *StagesStorageCacheHttpHandlerLegacy) handleGetStagesBySignature() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetStagesBySignatureRequestLegacy
		var response GetStagesBySignatureResponseLegacy
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- GetStagesBySignature request %#v\n", request)
			found, stages, err := handler.StagesStorageCache.GetStagesByDigest(context.Background(), request.ProjectName, request.Signature)
			for _, s := range stages {
				response.Stages = append(response.Stages, StageIDLegacy{Signature: s.Digest, UniqueID: s.UniqueID})
			}
			response.Found = found
			response.Err.Error = err
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- GetStagesBySignature response %#v\n", response)
		})
	}
}

type StoreStagesBySignatureRequestLegacy struct {
	ProjectName string          `json:"projectName"`
	Signature   string          `json:"signature"`
	Stages      []StageIDLegacy `json:"stages"`
}

type StoreStagesBySignatureResponseLegacy struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandlerLegacy) handleStoreStagesBySignature() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request StoreStagesBySignatureRequestLegacy
		var response StoreStagesBySignatureResponseLegacy
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- StoreStagesBySignature request %#v\n", request)
			var stages []image.StageID
			for _, s := range request.Stages {
				stages = append(stages, *image.NewStageID(s.Signature, s.UniqueID))
			}
			response.Err.Error = handler.StagesStorageCache.StoreStagesByDigest(context.Background(), request.ProjectName, request.Signature, stages)
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- StoreStagesBySignature response %#v\n", response)
		})
	}
}

type DeleteStagesBySignatureRequestLegacy struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}

type DeleteStagesBySignatureResponseLegacy struct {
	Err util.SerializableError `json:"err"`
}

func (handler *StagesStorageCacheHttpHandlerLegacy) handleDeleteStagesBySignature() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteStagesBySignatureRequestLegacy
		var response DeleteStagesBySignatureResponseLegacy
		HandleRequest(w, r, &request, &response, func() {
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- DeleteStagesBySignature request %#v\n", request)
			response.Err.Error = handler.StagesStorageCache.DeleteStagesByDigest(context.Background(), request.ProjectName, request.Signature)
			logboek.Debug().LogF("StagesStorageCacheHttpHandlerLegacy -- DeleteStagesBySignature response %#v\n", response)
		})
	}
}
