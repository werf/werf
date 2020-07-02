package synchronization_server

import (
	"fmt"
	"net/http"

	"github.com/werf/werf/pkg/storage"
)

func NewLockManagerHttpClient(url string) *LockManagerHttpClient {
	return &LockManagerHttpClient{
		URL:        url,
		HttpClient: &http.Client{},
	}
}

type LockManagerHttpClient struct {
	URL        string
	HttpClient *http.Client
}

func (client *LockManagerHttpClient) LockStage(projectName, signature string) (storage.LockHandle, error) {
	var request = LockStageRequest{projectName, signature}
	var response LockStageResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "lock-stage"), request, &response); err != nil {
		return storage.LockHandle{}, err
	}
	return response.LockHandle, response.Err.Error
}

func (client *LockManagerHttpClient) LockStageCache(projectName, signature string) (storage.LockHandle, error) {
	var request = LockStageCacheRequest{projectName, signature}
	var response LockStageCacheResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "lock-stage-cache"), request, &response); err != nil {
		return storage.LockHandle{}, err
	}
	return response.LockHandle, response.Err.Error
}

func (client *LockManagerHttpClient) LockImage(projectName, imageName string) (storage.LockHandle, error) {
	var request = LockImageRequest{projectName, imageName}
	var response LockImageResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "lock-image"), request, &response); err != nil {
		return storage.LockHandle{}, err
	}
	return response.LockHandle, response.Err.Error
}

func (client *LockManagerHttpClient) LockStagesAndImages(projectName string, opts storage.LockStagesAndImagesOptions) (storage.LockHandle, error) {
	var request = LockStagesAndImagesRequest{projectName, opts}
	var response LockStagesAndImagesResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "lock-stages-and-images"), request, &response); err != nil {
		return storage.LockHandle{}, err
	}
	return response.LockHandle, response.Err.Error
}

func (client *LockManagerHttpClient) LockDeployProcess(projectName string, releaseName string, kubeContextName string) (storage.LockHandle, error) {
	var request = LockDeployProcessRequest{projectName, releaseName, kubeContextName}
	var response LockDeployProcessResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "lock-deploy-process"), request, &response); err != nil {
		return storage.LockHandle{}, err
	}
	return response.LockHandle, response.Err.Error
}

func (client *LockManagerHttpClient) Unlock(lockHandle storage.LockHandle) error {
	var request = UnlockRequest{lockHandle}
	var response UnlockResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "unlock"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}
