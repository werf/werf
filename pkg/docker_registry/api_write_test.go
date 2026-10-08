package docker_registry

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Registry publication completion", func() {
	ginkgo.DescribeTable("preserves a manifest upload failure", func(ctx ginkgo.SpecContext, index, wrapped bool) {
		fixture := newWritableBearerRegistryFixture()
		ginkgo.DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		registry.httpTransport = bearerRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/manifests/") {
				if wrapped {
					return nil, fmt.Errorf("upload manifest: %w", io.EOF)
				}
				return nil, io.EOF
			}
			return remote.DefaultTransport.RoundTrip(req)
		})
		ref, err := name.ParseReference(strings.TrimPrefix(fixture.server.URL, "http://") + "/repo:tag")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		var imageOrIndex any = empty.Image
		if index {
			imageOrIndex = empty.Index
		}
		err = registry.writeToRemote(ctx, ref, imageOrIndex)
		gomega.Expect(err).To(gomega.MatchError(io.EOF))
		_, err = registry.GetRepoImage(ctx, ref.Name())
		gomega.Expect(IsImageNotFoundError(err)).To(gomega.BeTrue(), "the failed push must not publish the tag: %v", err)
	},
		ginkgo.Entry("image EOF", false, false),
		ginkgo.Entry("image wrapped EOF", false, true),
		ginkgo.Entry("index EOF", true, false),
		ginkgo.Entry("index wrapped EOF", true, true),
	)

	ginkgo.DescribeTable("waits for a successful manifest upload", func(ctx ginkgo.SpecContext, index bool) {
		fixture := newWritableBearerRegistryFixture()
		ginkgo.DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		ref, err := name.ParseReference(strings.TrimPrefix(fixture.server.URL, "http://") + "/repo:tag")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var imageOrIndex any = empty.Image
		if index {
			imageOrIndex = empty.Index
		}

		gomega.Expect(registry.writeToRemote(ctx, ref, imageOrIndex)).To(gomega.Succeed())
		info, err := registry.GetRepoImage(ctx, ref.Name())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.IsIndex).To(gomega.Equal(index))
	},
		ginkgo.Entry("image", false),
		ginkgo.Entry("index", true),
	)
})
