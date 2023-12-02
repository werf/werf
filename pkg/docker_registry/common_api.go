package docker_registry

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/werf/logboek"
	transport2 "github.com/werf/werf/pkg/docker_registry/transport"
)

type apiError struct {
	error
}

type doRequestOptions struct {
	Headers       map[string]string
	BasicAuth     doRequestBasicAuth
	AcceptedCodes []int
	SkipTlsVerify bool
}

type doRequestBasicAuth struct {
	username string
	password string
}

func doRequest(ctx context.Context, method, url string, body io.Reader, options doRequestOptions) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, err
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	if options.BasicAuth.username != "" && options.BasicAuth.password != "" {
		req.SetBasicAuth(options.BasicAuth.username, options.BasicAuth.password)
	}

	logboek.Context(ctx).Debug().LogF("--> %s %s\n", method, url)
	resp, err := getHTTPClient(options.SkipTlsVerify).Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, respBody, err
	}
	logboek.Context(ctx).Debug().LogF("<-- %s %s\n", resp.Status, string(respBody))

	if err := transport.CheckError(resp, options.AcceptedCodes...); err != nil {
		errMsg := err.Error()
		if len(respBody) != 0 {
			errMsg += fmt.Sprintf(" (body: %s)", strings.TrimSpace(string(respBody)))
		}

		return resp, respBody, errors.New(errMsg)
	}

	return resp, respBody, nil
}

func getHTTPClient(skipTlsVerify bool) *http.Client {
	return &http.Client{
		Transport: getHttpTransport(skipTlsVerify),
	}
}

func getHttpTransport(skipTlsVerify bool) http.RoundTripper {
	t := remote.DefaultTransport.(*http.Transport).Clone()

	if skipTlsVerify {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		t.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	}

	return transport.NewRetry(transport2.NewRetryAfter(t))
}
