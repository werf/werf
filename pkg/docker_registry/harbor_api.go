package docker_registry

import (
	"context"
	"net/http"
	neturl "net/url"
	"path"
)

type harborApi struct{}

func newHarborApi() harborApi {
	return harborApi{}
}

func (api *harborApi) DeleteRepository(ctx context.Context, hostname, repository, username, password string) (*http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repositories", repository)
	url := u.String()

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
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
