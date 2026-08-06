package docker_registry

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

type fakeRegistry struct {
	Interface
}

func (r *fakeRegistry) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	return []string{"latest"}, nil
}

var _ = Describe("timingDockerRegistry", func() {
	It("records a registry API measurement into the collector from the context", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		registry := newTimingDockerRegistry(&fakeRegistry{})

		tags, err := registry.Tags(ctx, "repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(tags).To(Equal([]string{"latest"}))

		summary := collector.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Operation).To(Equal(opstats.Operation("registry: Tags")))
		Expect(summary[0].Count).To(Equal(1))
	})
})
