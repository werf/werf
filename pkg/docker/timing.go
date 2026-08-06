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
// collector, so the wrapper is safe to install unconditionally. It never
// fails: when the client cannot be rebuilt, the original unmeasured client is
// returned — diagnostics must not break the build.
//
// A non-*http.Transport leaves the client's baseTransport unset, which
// disables the transport-derived dialer and TLS config used by hijacked
// connections (attach/exec). Only unix and npipe hosts survive that via the
// proto-based dialing fallback, so other hosts (tcp, ssh) are left unwrapped
// rather than broken.
func wrapAPIClientTransport(apiClient client.APIClient, httpHeaders map[string]string) client.APIClient {
	concreteClient, ok := apiClient.(*client.Client)
	if !ok {
		return apiClient
	}

	host := concreteClient.DaemonHost()
	if !strings.HasPrefix(host, "unix://") && !strings.HasPrefix(host, "npipe://") {
		return apiClient
	}

	httpClient := concreteClient.HTTPClient()
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	httpClient.Transport = &timingRoundTripper{next: httpClient.Transport}

	opts := []client.Opt{
		client.WithHost(host),
		client.WithHTTPClient(httpClient),
		client.WithUserAgent(command.UserAgent()),
		client.WithVersionFromEnv(),
		client.WithAPIVersionNegotiation(),
	}
	if len(httpHeaders) > 0 {
		opts = append(opts, client.WithHTTPHeaders(httpHeaders))
	}

	wrappedClient, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return apiClient
	}

	return wrappedClient
}
