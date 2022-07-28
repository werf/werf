package docker_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"path"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
)

type selectelApi struct{}

func newSelectelApi() selectelApi {
	return selectelApi{}
}

func (api *selectelApi) deleteRepository(ctx context.Context, hostname, registryId, repository, token string) (*http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api/v1")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repositories", repository)
	q := u.Query()
	q.Set("registryId", registryId)
	u.RawQuery = q.Encode()
	url := u.String()

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusNoContent},
	})

	return resp, err
}

func (api *selectelApi) deleteReference(ctx context.Context, hostname, registryId, repository, reference, token string) (*http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api/v1")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repositories", repository, reference)
	q := u.Query()
	q.Set("registryId", registryId)
	u.RawQuery = q.Encode()
	url := u.String()

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusNoContent},
	})

	return resp, err
}

func (api *selectelApi) getToken(ctx context.Context, username, password, account, vpc string) (string, error) {

	identityUrl := "https://api.selvpc.ru/identity/v3"

	scope := gophercloud.AuthScope{
		DomainName:  account,
		ProjectName: vpc,
	}

	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: identityUrl,
		Username:         username,
		Password:         password,
		DomainName:       account,
		Scope:            &scope,
		AllowReauth:      true,
	}

	provider, err := openstack.AuthenticatedClient(authOptions)
	if err != nil {
		return "", err
	}

	client, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return "", err
	}

	resp := tokens.Create(client, &authOptions)
	token, err := resp.ExtractTokenID()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (api *selectelApi) getRegistryId(ctx context.Context, hostname, registry, token string) (string, *http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api/v1/registries")
	if err != nil {
		return "", nil, err
	}

	resp, respBody, err := doRequest(ctx, http.MethodGet, u.String(), nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusOK},
	})

	if err != nil {
		return "", resp, err
	}

	resBodyJson := []map[string]interface{}{}
	if err := json.Unmarshal(respBody, &resBodyJson); err != nil {
		return "", resp, err
	}

	for _, repo := range resBodyJson {
		registryName, okName := repo["name"].(string)
		registryID, okID := repo["id"].(string)
		if !okID || !okName || registryName == "" {
			return "", resp, fmt.Errorf("unexpected selectel api response body: %s", string(respBody))
		}
		if registryName == registry {
			return registryID, resp, nil
		}
	}

	return "", resp, fmt.Errorf("unexpected selectel api response body: %s", string(respBody))

}

func (api *selectelApi) getTags(ctx context.Context, hostname, registryId, repository, token string) ([]string, *http.Response, error) {
	u, err := neturl.Parse("https://" + hostname + "/api/v1")
	if err != nil {
		return nil, nil, err
	}
	u.Path = path.Join(u.Path, "repositories", repository, "tags")

	resp, respBody, err := doRequest(ctx, http.MethodGet, u.String(), nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusOK},
	})

	if err != nil {
		return nil, resp, err
	}

	tagsBodyJson := []string{}
	if err := json.Unmarshal(respBody, &tagsBodyJson); err != nil {
		return nil, resp, err
	}

	return tagsBodyJson, resp, nil
}
