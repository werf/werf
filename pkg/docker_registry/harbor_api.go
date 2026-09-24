package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"path"
	"strings"
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
	project, repositoryName, ok := strings.Cut(repository, "/")
	if !ok || project == "" || repositoryName == "" {
		return nil, fmt.Errorf("invalid Harbor repository %q: expected project/name", repository)
	}

	u, err := neturl.Parse("https://" + hostname)
	if err != nil {
		return nil, err
	}

	options := doRequestOptions{
		Headers: map[string]string{
			"Accept": "application/json",
		},
		BasicAuth: doRequestBasicAuth{
			username: username,
			password: password,
		},
		AcceptedCodes: []int{http.StatusOK, http.StatusAccepted},
	}

	u.Path = path.Join(u.Path, "api", "v2.0", "projects", project, "repositories", neturl.PathEscape(repositoryName))
	resp, _, err := doRequest(ctx, api.httpClient, http.MethodDelete, u.String(), nil, options)
	if err == nil || resp == nil || resp.StatusCode != http.StatusNotFound {
		return resp, err
	}

	u.Path = path.Join("api", "repositories", repository)
	resp, _, err = doRequest(ctx, api.httpClient, http.MethodDelete, u.String(), nil, options)
	return resp, err
}
