package bundles

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = DescribeTable("normalizeAddr",
	func(addr, expected string) {
		normalized, err := normalizeAddr(addr)
		Expect(err).To(Succeed())
		Expect(normalized).To(Equal(expected))
	},
	Entry("Docker Hub user repo", "user/app", "docker.io/user/app"),
	Entry("Docker Hub user repo with a tag", "user/app:1.0.0+build", "docker.io/user/app:1.0.0+build"),
	Entry("Docker Hub official repo", "app:latest", "docker.io/library/app:latest"),
	Entry("Docker Hub repo with the schema", "docker://user/app:1.0.0", "docker://docker.io/user/app:1.0.0"),
	Entry("Docker Hub repo with a digest", "user/app@sha256:"+sha256Hex, "docker.io/user/app@sha256:"+sha256Hex),
	Entry("Docker Hub explicit host", "docker.io/user/app", "docker.io/user/app"),
	Entry("Docker Hub index host", "index.docker.io/user/app:1", "docker.io/user/app:1"),
	Entry("registry with a port", "localhost:5000/app:1.0.0", "localhost:5000/app:1.0.0"),
	Entry("registry with a port without a tag", "localhost:5000/app", "localhost:5000/app"),
	Entry("registry host", "registry.example.com/group/app:v1", "registry.example.com/group/app:v1"),
	Entry("archive", "archive:user/app.tar.gz", "archive:user/app.tar.gz"),
)

const sha256Hex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

var _ = DescribeTable("Docker Hub short repository names",
	func(ctx0 context.Context, run func(ctx context.Context, repo string, client *registry.Client) error) {
		ctx := logging.WithLogger(ctx0)
		transport := &hostRecordingTransport{}
		client, err := registry.NewClient(
			registry.ClientOptCredentialsFile(testCredentialsPath()),
			registry.ClientOptHTTPClient(&http.Client{Transport: transport}),
		)
		Expect(err).To(Succeed())

		Expect(run(ctx, "user/app", client)).NotTo(Succeed())
		Expect(transport.hosts).NotTo(BeEmpty())
		Expect(transport.hosts).To(HaveEach("registry-1.docker.io"))
	},
	Entry("publish", func(ctx context.Context, repo string, client *registry.Client) error {
		return Publish(ctx, writeTestChart(), repo, "1.0.0", client, PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})
	}),
	Entry("pull", func(ctx context.Context, repo string, client *registry.Client) error {
		return Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), client, PullOptions{})
	}),
)

type hostRecordingTransport struct {
	mu    sync.Mutex
	hosts []string
}

func (t *hostRecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.hosts = append(t.hosts, req.URL.Host)

	return nil, errors.New("no network in tests")
}
