package docker_registry

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/werf/logboek"
)

type apiError struct {
	error
}

type doRequestOptions struct {
	Headers       map[string]string
	BasicAuth     doRequestBasicAuth
	AcceptedCodes []int
}

type doRequestBasicAuth struct {
	username string
	password string
}

func doRequest(method, url string, body io.Reader, options doRequestOptions) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	if options.BasicAuth.username != "" && options.BasicAuth.password != "" {
		req.SetBasicAuth(options.BasicAuth.username, options.BasicAuth.password)
	}

	logboek.Debug.LogF("--> %s %s\n", method, url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, respBody, err
	}
	logboek.Debug.LogF("<-- %s %s\n", resp.Status, string(respBody))

	if err := transport.CheckError(resp, options.AcceptedCodes...); err != nil {
		errMsg := err.Error()
		if len(respBody) != 0 {
			errMsg += fmt.Sprintf(" (body: %s)", strings.TrimSpace(string(respBody)))
		}

		return resp, respBody, errors.New(errMsg)
	}

	return resp, respBody, nil
}
