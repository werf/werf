package docker

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/cli/cli/connhelper/commandconn"
	"github.com/moby/moby/client"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type failingTransport struct {
	err error
}

var _ http.RoundTripper = (*failingTransport)(nil)

func (t *failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.err
}

func daemonTransportContext(transportErr error) context.Context {
	api, err := client.New(client.WithHost("http://127.0.0.1:1"), client.WithHTTPClient(&http.Client{Transport: &failingTransport{err: transportErr}}))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}

func sshDaemonContext(diagnostic string) context.Context {
	if runtime.GOOS == "windows" {
		ginkgo.Skip("SSH connection-helper fixture requires a POSIX shell")
	}
	diagnosticPath := filepath.Join(ginkgo.GinkgoT().TempDir(), "stderr")
	gomega.Expect(os.WriteFile(diagnosticPath, []byte(diagnostic+"\n"), 0o600)).To(gomega.Succeed())
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return commandconn.New(ctx, "/bin/sh", "-c", `read request; cat "$0" >&2; exit 255`, diagnosticPath)
	}}
	api, err := client.New(client.WithHost("http://docker.example"), client.WithHTTPClient(&http.Client{Transport: transport}))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}

func daemonSettingsContext(handler http.Handler) context.Context {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	api, err := client.New(client.WithHost(server.URL))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}
