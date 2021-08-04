package docker_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const gitHubGraphqlAPIUrl = "https://api.github.com/graphql"

type gitHubGraphqlApi struct{}

func newGitHubGraphqlApi() gitHubGraphqlApi {
	return gitHubGraphqlApi{}
}

func (api *gitHubGraphqlApi) deletePackageVersion(ctx context.Context, packageVersionId, token string) (*http.Response, error) {
	body := []byte(fmt.Sprintf(`{"query":"mutation { deletePackageVersion(input:{packageVersionId:\"%s\"}) { success }}"}"}`, packageVersionId))

	resp, _, err := api.doRequest(ctx, http.MethodPost, gitHubGraphqlAPIUrl, bytes.NewBuffer(body), doRequestOptions{
		Headers: map[string]string{
			"Accept":        "application/vnd.github.package-deletes-preview+json",
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted},
	})

	return resp, err
}

func (api *gitHubGraphqlApi) getPackageVersionId(ctx context.Context, owner, repo, packageName, versionName, token string) (string, *http.Response, error) {
	body := []byte(fmt.Sprintf(`{"query":"query{repository(owner:\"%s\",name:\"%s\"){packages(names: \"%s\", first: 1){nodes{id, version(version: \"%s\") { id, version }}}}}"}`, owner, repo, packageName, versionName))

	resp, respBody, err := api.doRequest(ctx, http.MethodPost, gitHubGraphqlAPIUrl, bytes.NewBuffer(body), doRequestOptions{
		Headers: map[string]string{
			"Accept":        "application/vnd.github.packages-preview+json",
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted},
	})
	if err != nil {
		return "", resp, err
	}

	respJson := &struct {
		Data struct {
			Repository struct {
				Packages struct {
					Nodes []struct {
						Id      string
						Version struct {
							Version string
							Id      string
						}
					}
				}
			}
		}
	}{}

	if err := json.Unmarshal(respBody, &respJson); err != nil {
		return "", resp, fmt.Errorf("unexpected body %s", string(respBody))
	}

	nodes := respJson.Data.Repository.Packages.Nodes
	if len(nodes) != 1 || nodes[0].Version.Id == "" {
		return "", nil, fmt.Errorf("unexpected body %s", string(respBody))
	}

	return nodes[0].Version.Id, resp, nil
}

func (api *gitHubGraphqlApi) doRequest(ctx context.Context, method, url string, body io.Reader, options doRequestOptions) (*http.Response, []byte, error) {
	resp, respBody, err := doRequest(ctx, method, url, body, options)
	if err != nil {
		return resp, respBody, err
	}

	respBodyJson := map[string]interface{}{}
	err = json.Unmarshal(respBody, &respBodyJson)
	if err != nil {
		return resp, respBody, err
	}

	_, ok := respBodyJson["errors"]
	if ok {
		return resp, respBody, errors.New(string(respBody))
	}

	return resp, respBody, nil
}
