package synchronization_server

import (
	"net/http"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/storage"
)

func NewLockManagerHttpHandler(lockManager storage.LockManager) *LockManagerHttpHandler {
	handler := &LockManagerHttpHandler{
		LockManager: lockManager,
		ServeMux:    http.NewServeMux(),
	}
	handler.HandleFunc("/lock-stage", handler.handleLockStage)
	handler.HandleFunc("/lock-stage-cache", handler.handleLockStageCache)
	handler.HandleFunc("/lock-image", handler.handleLockImage)
	handler.HandleFunc("/lock-stages-and-images", handler.handleLockStagesAndImages)
	handler.HandleFunc("/lock-deploy-process", handler.handleLockDeployProcess)
	handler.HandleFunc("/unlock", handler.handleUnlock)

	return handler
}

type LockManagerHttpHandler struct {
	*http.ServeMux
	LockManager storage.LockManager
}

type LockStageRequest struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}
type LockStageResponse struct {
	Err        util.SerializableError `json:"err"`
	LockHandle storage.LockHandle     `json:"lockHandle"`
}

func (handler *LockManagerHttpHandler) handleLockStage(w http.ResponseWriter, r *http.Request) {
	var request LockStageRequest
	var response LockStageResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStage request %#v\n", request)
		response.LockHandle, response.Err.Error = handler.LockManager.LockStage(request.ProjectName, request.Signature)
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStage response %#v\n", response)
	})
}

type LockStageCacheRequest struct {
	ProjectName string `json:"projectName"`
	Signature   string `json:"signature"`
}
type LockStageCacheResponse struct {
	Err        util.SerializableError `json:"err"`
	LockHandle storage.LockHandle     `json:"lockHandle"`
}

func (handler *LockManagerHttpHandler) handleLockStageCache(w http.ResponseWriter, r *http.Request) {
	var request LockStageCacheRequest
	var response LockStageCacheResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStageCache request %#v\n", request)
		response.LockHandle, response.Err.Error = handler.LockManager.LockStageCache(request.ProjectName, request.Signature)
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStageCache response %#v\n", response)
	})
}

type LockImageRequest struct {
	ProjectName string `json:"projectName"`
	ImageName   string `json:"imageName"`
}
type LockImageResponse struct {
	Err        util.SerializableError `json:"err"`
	LockHandle storage.LockHandle     `json:"lockHandle"`
}

func (handler *LockManagerHttpHandler) handleLockImage(w http.ResponseWriter, r *http.Request) {
	var request LockImageRequest
	var response LockImageResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- LockImage request %#v\n", request)
		response.LockHandle, response.Err.Error = handler.LockManager.LockImage(request.ProjectName, request.ImageName)
		logboek.Debug.LogF("LockManagerHttpHandler -- LockImage response %#v\n", response)
	})
}

type LockStagesAndImagesRequest struct {
	ProjectName string                             `json:"projectName"`
	Opts        storage.LockStagesAndImagesOptions `json:"opts"`
}
type LockStagesAndImagesResponse struct {
	Err        util.SerializableError `json:"err"`
	LockHandle storage.LockHandle     `json:"lockHandle"`
}

func (handler *LockManagerHttpHandler) handleLockStagesAndImages(w http.ResponseWriter, r *http.Request) {
	var request LockStagesAndImagesRequest
	var response LockStagesAndImagesResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStagesAndImages request %#v\n", request)
		response.LockHandle, response.Err.Error = handler.LockManager.LockStagesAndImages(request.ProjectName, request.Opts)
		logboek.Debug.LogF("LockManagerHttpHandler -- LockStagesAndImages response %#v\n", response)
	})
}

type LockDeployProcessRequest struct {
	ProjectName     string `json:"projectName"`
	ReleaseName     string `json:"releaseName"`
	KubeContextName string `json:"kubeContextName"`
}
type LockDeployProcessResponse struct {
	Err        util.SerializableError `json:"err"`
	LockHandle storage.LockHandle     `json:"lockHandle"`
}

func (handler *LockManagerHttpHandler) handleLockDeployProcess(w http.ResponseWriter, r *http.Request) {
	var request LockDeployProcessRequest
	var response LockDeployProcessResponse
	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- LockDeployProcess request %#v\n", request)
		response.LockHandle, response.Err.Error = handler.LockManager.LockDeployProcess(request.ProjectName, request.ReleaseName, request.KubeContextName)
		logboek.Debug.LogF("LockManagerHttpHandler -- LockDeployProcess response %#v\n", response)
	})
}

type UnlockRequest struct {
	LockHandle storage.LockHandle `json:"lockHandle"`
}
type UnlockResponse struct {
	Err util.SerializableError `json:"err"`
}

func (handler *LockManagerHttpHandler) handleUnlock(w http.ResponseWriter, r *http.Request) {
	var request UnlockRequest
	var response UnlockResponse

	HandleRequest(w, r, &request, &response, func() {
		logboek.Debug.LogF("LockManagerHttpHandler -- Unlock request %#v\n", request)
		response.Err.Error = handler.LockManager.Unlock(request.LockHandle)
		logboek.Debug.LogF("LockManagerHttpHandler -- Unlock response %#v\n", response)
	})
}
