package docker

import (
	"net/http"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"

	"github.com/werf/werf/v2/pkg/opstats"
)

var _ http.RoundTripper = (*timingRoundTripper)(nil)

type timingRoundTripper struct {
	next http.RoundTripper
}

func (t *timingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer opstats.Observe(req.Context(), opstats.OperationDockerDaemon)()
	return t.next.RoundTrip(req)
}

// wrapAPIClientTransport rebuilds the API client around the same HTTP
// transport (preserving TLS and dialer setup) wrapped with per-request timing.
// Timing is recorded only when the request context carries an opstats
// collector, so the wrapper is safe to install unconditionally.
//
// A non-*http.Transport leaves the client's baseTransport unset, which
// disables the transport-derived dialer and TLS config used by hijacked
// connections (attach/exec). Only unix and npipe hosts survive that via the
// proto-based dialing fallback, so other hosts (tcp, ssh) are left unwrapped
// rather than broken.
func wrapAPIClientTransport(apiClient client.APIClient) (client.APIClient, error) {
	concreteClient, ok := apiClient.(*client.Client)
	if !ok {
		return apiClient, nil
	}

	host := concreteClient.DaemonHost()
	if !strings.HasPrefix(host, "unix://") && !strings.HasPrefix(host, "npipe://") {
		return apiClient, nil
	}

	httpClient := concreteClient.HTTPClient()
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	httpClient.Transport = &timingRoundTripper{next: httpClient.Transport}

	return client.NewClientWithOpts(
		client.WithHost(host),
		client.WithHTTPClient(httpClient),
		client.WithUserAgent(command.UserAgent()),
		client.WithVersionFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
}
