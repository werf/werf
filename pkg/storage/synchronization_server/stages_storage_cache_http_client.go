package synchronization_server

import (
	"context"
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

func (client *StagesStorageCacheHttpClient) GetAllStages(_ context.Context, projectName string) (bool, []image.StageID, error) {
	request := GetAllStagesRequest{projectName}
	var response GetAllStagesResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/v1/%s", client.URL, "get-all-stages"), request, &response); err != nil {
		return false, nil, err
	}
	return response.Found, response.Stages, response.Err.Error
}

func (client *StagesStorageCacheHttpClient) DeleteAllStages(_ context.Context, projectName string) error {
	request := DeleteAllStagesRequest{projectName}
	var response DeleteAllStagesResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/v1/%s", client.URL, "delete-all-stages"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}

func (client *StagesStorageCacheHttpClient) GetStagesByDigest(_ context.Context, projectName, digest string) (bool, []image.StageID, error) {
	request := GetStagesByDigestRequest{projectName, digest}
	var response GetStagesByDigestResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/v1/%s", client.URL, "get-stages-by-digest"), request, &response); err != nil {
		return false, nil, err
	}
	return response.Found, response.Stages, response.Err.Error
}

func (client *StagesStorageCacheHttpClient) StoreStagesByDigest(_ context.Context, projectName, digest string, stages []image.StageID) error {
	request := StoreStagesByDigestRequest{projectName, digest, stages}
	var response StoreStagesByDigestResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/v1/%s", client.URL, "store-stages-by-digest"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}

func (client *StagesStorageCacheHttpClient) DeleteStagesByDigest(_ context.Context, projectName, digest string) error {
	request := DeleteStagesByDigestRequest{projectName, digest}
	var response DeleteStagesByDigestResponse
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/v1/%s", client.URL, "delete-stages-by-digest"), request, &response); err != nil {
		return err
	}
	return response.Err.Error
}
