package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type quayApi struct{}

func newQuayApi() quayApi {
	return quayApi{}
}

func (api *quayApi) DeleteRepository(ctx context.Context, hostname, namespace, repository, token string) (*http.Response, error) {
	u, err := url.Parse("https://" + hostname + "/api/v1/")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repository", namespace, repository)

	reqUrl := u.String()
	reqAccept := "application/json"
	reqAuthorization := fmt.Sprintf("Bearer %s", token)

	resp, _, err := doRequest(ctx, http.MethodDelete, reqUrl, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":        reqAccept,
			"Authorization": reqAuthorization,
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent},
	})

	return resp, err
}
