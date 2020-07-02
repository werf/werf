package synchronization_server

import (
	"fmt"
	"net/http"

	"github.com/werf/werf/pkg/image"
)

func NewStagesStorageCacheHttpClient(url string) *StagesStorageCacheHttpClient {
	return &StagesStorageCacheHttpClient{
		URL:        url,
		HttpClient: &http.Client{},
	}
}

type StagesStorageCacheHttpClient struct {
	URL        string
	HttpClient *http.Client
}

func (client *StagesStorageCacheHttpClient) String() string {
	return fmt.Sprintf("http-client %s", client.URL)
}

func (client *StagesStorageCacheHttpClient) GetAllStages(projectName string) (bool, []image.StageID, error) {
	var request = GetAllStagesRequest{projectName}
	var response GetAllStagesResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "get-all-stages"), request, &response); err != nil {
		return false, nil, err
	}
	return response.Found, response.Stages, response.Err.Error
}

func (client *StagesStorageCacheHttpClient) DeleteAllStages(projectName string) error {
	var request = DeleteAllStagesRequest{projectName}
	var response DeleteAllStagesResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "delete-all-stages"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}

func (client *StagesStorageCacheHttpClient) GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error) {
	var request = GetStagesBySignatureRequest{projectName, signature}
	var response GetStagesBySignatureResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "get-stages-by-signature"), request, &response); err != nil {
		return false, nil, err
	}
	return response.Found, response.Stages, response.Err.Error
}

func (client *StagesStorageCacheHttpClient) StoreStagesBySignature(projectName, signature string, stages []image.StageID) error {
	var request = StoreStagesBySignatureRequest{projectName, signature, stages}
	var response StoreStagesBySignatureResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "store-stages-by-signature"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}

func (client *StagesStorageCacheHttpClient) DeleteStagesBySignature(projectName, signature string) error {
	var request = DeleteStagesBySignatureRequest{projectName, signature}
	var response DeleteStagesBySignatureResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "delete-stages-by-signature"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}
