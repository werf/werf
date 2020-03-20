package docker_registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/flant/logboek"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

type dockerHubApi struct{}

func newDockerHubApi() dockerHubApi {
	return dockerHubApi{}
}

func (r *dockerHubApi) deleteTag(repository, tag, token string) (*http.Response, error) {
	u, err := url.Parse(repository)
	if err != nil {
		return nil, err
	}

	project := path.Base(u.Path)
	account := path.Base(strings.TrimSuffix(u.Path, project))

	reqUrl := fmt.Sprintf(
		"https://hub.docker.com/v2/repositories/%s/%s/tags/%s/",
		account,
		project,
		tag,
	)
	reqAccept := "application/json"
	reqAuthorization := fmt.Sprintf("JWT %s", token)

	logboek.Debug.LogF("--> %s %s\n", http.MethodDelete, reqUrl)
	req, err := http.NewRequest(http.MethodDelete, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", reqAccept)
	req.Header.Set("Authorization", reqAuthorization)

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

	if resp.StatusCode == http.StatusForbidden {
		return resp, fmt.Errorf(
			"DELETE %s failed: %s",
			reqUrl,
			string(respBody),
		)
	}

	if err := transport.CheckError(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent); err != nil {
		return resp, err
	}

	return resp, nil
}

func (r *dockerHubApi) getToken(username, password string) (string, *http.Response, error) {
	values := map[string]string{
		"username": username,
		"password": password,
	}

	jsonValue, err := json.Marshal(values)
	if err != nil {
		return "", nil, err
	}

	reqUrl := "https://hub.docker.com/v2/users/login/"
	reqContentType := "application/json"
	reqBody := bytes.NewBuffer(jsonValue)

	logboek.Debug.LogF("--> %s %s\n", http.MethodPost, reqUrl)
	resp, err := http.Post(reqUrl, reqContentType, reqBody)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp, err
	}

	logboek.Debug.LogF("<-- %s %s\n", resp.Status, respBody)

	if err := transport.CheckError(resp, http.StatusOK, http.StatusAccepted, http.StatusUnauthorized); err != nil {
		return "", resp, err
	}

	resBodyJson := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &resBodyJson); err != nil {
		return "", resp, err
	}

	token, ok := resBodyJson["token"].(string)
	if !ok {
		return "", nil, fmt.Errorf("unexpected docker hub api response body: %s", string(respBody))
	}

	return token, resp, nil
}
