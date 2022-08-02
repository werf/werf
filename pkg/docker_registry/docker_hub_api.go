package docker_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type dockerHubApi struct{}

func newDockerHubApi() dockerHubApi {
	return dockerHubApi{}
}

func (api *dockerHubApi) deleteRepository(ctx context.Context, account, project, token string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://hub.docker.com/v2/repositories/%s/%s/",
		account,
		project,
	)

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("JWT %s", token),
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent},
	})

	return resp, err
}

func (api *dockerHubApi) deleteTag(ctx context.Context, account, project, tag, token string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://hub.docker.com/v2/repositories/%s/%s/tags/%s/",
		account,
		project,
		tag,
	)

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("JWT %s", token),
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent},
	})

	return resp, err
}

func (api *dockerHubApi) getToken(ctx context.Context, username, password string) (string, *http.Response, error) {
	url := "https://hub.docker.com/v2/users/login/"
	body, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return "", nil, err
	}

	resp, respBody, err := doRequest(ctx, http.MethodPost, url, bytes.NewBuffer(body), doRequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted},
	})
	if err != nil {
		return "", resp, err
	}

	resBodyJson := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &resBodyJson); err != nil {
		return "", resp, err
	}

	token, ok := resBodyJson["token"].(string)
	if !ok || token == "" {
		return "", resp, fmt.Errorf("unexpected docker hub api response body: %s", string(respBody))
	}

	return token, resp, nil
}
