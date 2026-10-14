package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Publication cache wiring", func() {
	for _, canceled := range []bool{false, true} {
		ginkgo.DescribeTable(fmt.Sprintf("exposes only successful writes to cached lookups (canceled=%t)", canceled), func(ctx ginkgo.SpecContext, operation string) {
			var listings atomic.Int32
			backend := registry.New()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/tags/list") {
					listings.Add(1)
				}
				backend.ServeHTTP(w, r)
			}))
			ginkgo.DeferCleanup(server.Close)
			repository := strings.TrimPrefix(server.URL, "http://") + "/repo"
			client, err := NewDockerRegistry(ctx, repository, DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(client.PushImage(ctx, repository+":source", &PushImageOptions{})).To(gomega.Succeed())
			source, err := client.GetRepoImage(ctx, repository+":source")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			before, err := client.Tags(ctx, repository, WithCachedTags())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(before).To(gomega.Equal([]string{"source"}))
			callCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			if canceled {
				cancel()
			}
			switch operation {
			case "push":
				err = client.PushImage(callCtx, repository+":destination", &PushImageOptions{})
			case "tag":
				err = client.TagRepoImage(callCtx, source, "destination")
			case "copy":
				err = client.CopyImage(callCtx, repository+":source", repository+":destination", CopyImageOptions{})
			case "mutate":
				err = client.MutateAndPushImage(callCtx, repository+":source", repository+":destination")
			}
			if canceled {
				gomega.Expect(err).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}
			after, err := client.Tags(ctx, repository, WithCachedTags())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			if canceled {
				gomega.Expect(after).To(gomega.Equal([]string{"source"}))
			} else {
				gomega.Expect(after).To(gomega.ConsistOf("source", "destination"))
			}
			gomega.Expect(listings.Load()).To(gomega.Equal(int32(1)))
		}, ginkgo.Entry("push", "push"), ginkgo.Entry("tag", "tag"), ginkgo.Entry("copy", "copy"), ginkgo.Entry("mutate", "mutate"))
	}
})
