package synchronization_server

import (
	"fmt"
	"net/http"
)

type SynchronizationClient struct {
	HttpClient *http.Client
	URL        string
}

func NewSynchronizationClient(url string) *SynchronizationClient {
	return &SynchronizationClient{
		URL:        url,
		HttpClient: &http.Client{},
	}
}

func (client *SynchronizationClient) NewClientID() (string, error) {
	request := NewClientIDRequest{}
	response := NewClientIDResponse{}
	if err := PerformPost(client.HttpClient, fmt.Sprintf("%s/%s", client.URL, "new-client-id"), request, &response); err != nil {
		return "", err
	}
	return response.ClientID, response.Err.Error
}
