package docker_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	neturl "net/url"
	"path"
	"strconv"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"

	"github.com/werf/logboek"
	parallelConstant "github.com/werf/werf/pkg/util/parallel/constant"
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

	resp, _, err := api.doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
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

	resp, _, err := api.doRequest(ctx, http.MethodDelete, url, nil, doRequestOptions{
		Headers: map[string]string{
			"Accept":       "application/json",
			"X-Auth-Token": token,
		},
		AcceptedCodes: []int{http.StatusNoContent},
	})

	return resp, err
}

func (api *selectelApi) getToken(ctx context.Context, username, password, account, vpc, vpcID string) (string, error) {
	var scope gophercloud.AuthScope

	const identityUrl = "https://api.selvpc.ru/identity/v3"

	if vpcID != "" {
		scope = gophercloud.AuthScope{
			ProjectID: vpcID,
		}
	} else {
		scope = gophercloud.AuthScope{
			ProjectName: vpc,
			DomainName:  account,
		}
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

	resp, respBody, err := api.doRequest(ctx, http.MethodGet, url, nil, doRequestOptions{
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
		AcceptedCodes: []int{http.StatusOK, http.StatusNotFound},
	})
	if err != nil {
		return nil, resp, err
	}

	tagsBodyJson := []string{}
	if resp.StatusCode == http.StatusNotFound {
		return tagsBodyJson, resp, nil
	}

	if err := json.Unmarshal(respBody, &tagsBodyJson); err != nil {
		return nil, resp, err
	}

	return tagsBodyJson, resp, nil
}

func (api *selectelApi) doRequest(ctx context.Context, method, url string, body io.Reader, options doRequestOptions) (*http.Response, []byte, error) {
	var seconds int
	resp, respBody, err := doRequest(ctx, method, url, body, options)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if resp.Header.Get("Retry-After") != "" {
				secondsString := resp.Header.Get("Retry-After")
				seconds, _ = strconv.Atoi(secondsString)
			}

			sleepSeconds := seconds + rand.Intn(15) + 5
			workerId := ctx.Value(parallelConstant.CtxBackgroundTaskIDKey)
			if workerId != nil {
				logboek.Context(ctx).Warn().LogF(
					"WARNING: Rate limit error occurred. Waiting for %d before retrying request... (worker %d).\nThe --parallel ($WERF_PARALLEL) and --parallel-tasks-limit ($WERF_PARALLEL_TASKS_LIMIT) options can be used to regulate parallel tasks.\n",
					sleepSeconds,
					workerId.(int),
				)
				logboek.Context(ctx).Warn().LogLn()
			} else {
				logboek.Context(ctx).Warn().LogF(
					"WARNING: Rate limit error occurred. Waiting for %d before retrying request...\n",
					sleepSeconds,
				)
			}

			time.Sleep(time.Second * time.Duration(sleepSeconds))
			return api.doRequest(ctx, method, url, body, options)
		}

		return resp, respBody, err
	}

	return resp, respBody, nil
}
