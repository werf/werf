package docker_registry

import (
	"context"
	"net/http"
	neturl "net/url"
	"path"
)

type harborApi struct {
	httpClient *http.Client
}

func newHarborApi() harborApi {
	return harborApi{
		httpClient: &http.Client{
			Transport: newHttpTransport(false),
		},
	}
}

func (api *harborApi) DeleteRepository(ctx context.Context, hostname, repository, username, password string) (*http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repositories", repository)
	url := u.String()

	resp, _, err := doRequest(ctx, api.httpClient, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept": "application/json",
		},
		BasicAuth: doRequestBasicAuth{
			username: username,
			password: password,
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted},
	})

	return resp, err
}
