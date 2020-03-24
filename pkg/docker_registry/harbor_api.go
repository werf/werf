package docker_registry

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/flant/logboek"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

type harborApi struct{}

func newHarborApi() harborApi {
	return harborApi{}
}

func (api *harborApi) DeleteRepository(hostname, repository, username, password string) (*http.Response, error) {
	u, err := url.Parse("https://" + hostname + "/api")
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "repositories", repository)

	reqUrl := u.String()
	reqAccept := "application/json"

	logboek.Debug.LogF("--> %s %s\n", http.MethodDelete, nil)
	req, err := http.NewRequest(http.MethodDelete, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", reqAccept)
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	logboek.Debug.LogF("<-- %s %s\n", resp.Status, respBody)

	if err := transport.CheckError(resp, http.StatusOK, http.StatusAccepted); err != nil {
		return resp, err
	}

	return resp, nil
}
