package publish

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/chartyaml"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/pkg/config"
)

var _ = Describe("chart pack options", func() {
	It("splits annotations, records the images and the release converge would deploy", func() {
		ctx := context.Background()

		prevCommonCmdData := commonCmdData
		DeferCleanup(func() { commonCmdData = prevCommonCmdData })

		commonCmdData.Environment = "production"
		commonCmdData.ExtraAnnotations = []string{"team=payments", "project.werf.io/x=y", "gitlab.ci.werf.io/pipeline=1", "werf.io/release-channel=beta"}
		commonCmdData.ExtraLabels = []string{"tier=backend"}

		namespace := "[[ project ]]-[[ env ]]-ns"
		werfConfig := &config.WerfConfig{Meta: &config.Meta{Project: "proj", Deploy: config.MetaDeploy{Namespace: &namespace}}}

		serviceValues := parseValues(`
global:
  werf:
    images:
      backend:
        ref: registry.example.com/proj:backend@sha256:aaa
        ref_tag: registry.example.com/proj:backend
        digest: sha256:aaa
      frontend:
        ref: registry.example.com/proj:frontend
        ref_tag: registry.example.com/proj:frontend
`)

		opts, err := newChartPackOptions(ctx, werfConfig, serviceValues)
		Expect(err).To(Succeed())

		serviceAnnotations := map[string]string{
			"project.werf.io/name":       "proj",
			"project.werf.io/env":        "production",
			"project.werf.io/x":          "y",
			"gitlab.ci.werf.io/pipeline": "1",
		}
		Expect(opts.ExtraAnnotations).To(Equal(map[string]string{"team": "payments"}))
		Expect(opts.ExtraRuntimeAnnotations).To(Equal(serviceAnnotations))
		Expect(opts.ReleaseInfoAnnotations).To(Equal(serviceAnnotations))
		Expect(opts.ExtraLabels).To(Equal(map[string]string{"tier": "backend"}))
		Expect(opts.ExtraValues).To(Equal(serviceValues))
		Expect(opts.Release).To(Equal(&werfdata.ReleaseMetadata{Name: "proj-production", Namespace: "proj-production-ns", Env: "production"}))
		Expect(opts.Images).To(Equal([]chartyaml.Image{
			{Name: "backend", Image: "registry.example.com/proj:backend@sha256:aaa"},
			{Name: "frontend", Image: "registry.example.com/proj:frontend"},
		}))
	})
})
