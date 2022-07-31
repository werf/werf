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

func (api *selectelApi) makeApiUrl(hostname, registryId string, pathParts ...string) (string, error) {
	u, err := neturl.Parse("https://" + hostname + "/api/v1")
	if err != nil {
		return "", err
	}

	parts := append([]string{u.Path}, pathParts...)
	u.Path = path.Join(parts...)

	if registryId != "" {
		q := u.Query()
		q.Set("registryId", registryId)
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

func (api *selectelApi) deleteRepository(ctx context.Context, hostname, registryId, repository, token string) (*http.Response, error) {
	url, err := api.makeApiUrl(hostname, registryId, "repositories", repository)
	if err != nil {
		return nil, err
	}

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
	url, err := api.makeApiUrl(hostname, registryId, "repositories", repository, reference)
	if err != nil {
		return nil, err
	}

	resp, _, err := doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusNoContent},
	})

	return resp, err
}

func (api *selectelApi) getToken(ctx context.Context, username, password, account, vpc, vpcID string) (string, error) {

	identityUrl := "https://api.selvpc.ru/identity/v3"

	scope := gophercloud.AuthScope{
		DomainName: account,
	}
	if vpcID != "" {
		scope.ProjectID = vpcID
	} else {
		scope.ProjectName = vpc
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
	url, err := api.makeApiUrl(hostname, "", "registries")
	if err != nil {
		return "", nil, err
	}

	resp, respBody, err := doRequest(ctx, http.MethodGet, url, nil, doRequestOptions{
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
	url, err := api.makeApiUrl(hostname, registryId, "repositories", repository, "tags")
	if err != nil {
		return nil, nil, err
	}

	resp, respBody, err := doRequest(ctx, http.MethodGet, url, nil, doRequestOptions{
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
