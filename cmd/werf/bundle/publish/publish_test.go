package publish

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/chartyaml"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
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
		Expect(opts.Release).To(Equal(&werfdata.ReleaseMetadata{Name: "proj-production", Namespace: "proj-production-ns"}))
		Expect(opts.Images).To(Equal([]chartyaml.Image{
			{Name: "backend", Image: "registry.example.com/proj:backend@sha256:aaa"},
			{Name: "frontend", Image: "registry.example.com/proj:frontend"},
		}))
	})
})

var _ = Describe("chart pack options of the default files switches", func() {
	DescribeTable("keep the switches stored in a committed .werf-data when not enabled",
		func(ctx context.Context, args []string, expected werfdata.Metadata) {
			prevCommonCmdData := commonCmdData
			DeferCleanup(func() { commonCmdData = prevCommonCmdData })

			chartDir := GinkgoT().TempDir()
			writeFile(filepath.Join(chartDir, "Chart.yaml"), "apiVersion: v2\nname: app\nversion: 1.0.0\n")
			writeFile(filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName), "apiVersion: v2\nvalues:\n  defaultValuesDisable: true\n  defaultSecretValuesDisable: true\ndefaultPatchesDisable: true\n")

			cmd := NewCmd(ctx)
			Expect(cmd.ParseFlags(args)).To(Succeed())

			opts, err := newChartPackOptions(ctx, &config.WerfConfig{Meta: &config.Meta{Project: "proj"}}, nil)
			Expect(err).To(Succeed())

			opts.Name = "app"
			opts.Version = "1.0.0"
			opts.Destination = GinkgoT().TempDir()

			result, err := action.ChartPack(ctx, chartDir, opts)
			Expect(err).To(Succeed())

			archive, err := os.ReadFile(result.ArchivePath)
			Expect(err).To(Succeed())

			loaded, err := loader.LoadArchive(nelmcommon.ContextWithHelmOptions(ctx, nelmcommon.HelmOptions{
				ChartLoadOpts: nelmcommon.ChartLoadOptions{NoSecrets: true},
			}), bytes.NewReader(archive))
			Expect(err).To(Succeed())

			data, err := werfdata.Parse(loaded.(*chart.Chart).WerfDataFiles)
			Expect(err).To(Succeed())
			Expect(data.Metadata.Values.DefaultValuesDisable).To(Equal(expected.Values.DefaultValuesDisable))
			Expect(data.Metadata.Values.DefaultSecretValuesDisable).To(Equal(expected.Values.DefaultSecretValuesDisable))
			Expect(data.Metadata.DefaultPatchesDisable).To(Equal(expected.DefaultPatchesDisable))
		},
		Entry("omitted",
			[]string{},
			werfdata.Metadata{Values: werfdata.ValuesMetadata{DefaultValuesDisable: true, DefaultSecretValuesDisable: true}, DefaultPatchesDisable: true}),
		Entry("--disable-default-values=false",
			[]string{"--disable-default-values=false"},
			werfdata.Metadata{Values: werfdata.ValuesMetadata{DefaultValuesDisable: true, DefaultSecretValuesDisable: true}, DefaultPatchesDisable: true}),
		Entry("--disable-default-secret-values=false",
			[]string{"--disable-default-secret-values=false"},
			werfdata.Metadata{Values: werfdata.ValuesMetadata{DefaultValuesDisable: true, DefaultSecretValuesDisable: true}, DefaultPatchesDisable: true}),
		Entry("--no-default-patches=false",
			[]string{"--no-default-patches=false"},
			werfdata.Metadata{Values: werfdata.ValuesMetadata{DefaultValuesDisable: true, DefaultSecretValuesDisable: true}, DefaultPatchesDisable: true}),
	)
})
