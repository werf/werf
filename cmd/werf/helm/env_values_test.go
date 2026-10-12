package helm

import (
	"context"
	"io"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/3p-helm-for-werf-helm/pkg/action"
	"github.com/werf/3p-helm-for-werf-helm/pkg/chart/loader"
	"github.com/werf/werf/v2/cmd/werf/common"
	chart_extender "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm"
)

var _ = ginkgo.Describe("Legacy Helm secret-values environment", func() {
	ginkgo.BeforeEach(clearListEnv)
	ginkgo.DescribeTable("loads environment secret files before explicit flags", func(bare, suffixed bool) {
		oldData, oldLoadOptions := templateCmdData, loader.GlobalLoadOptions
		ginkgo.DeferCleanup(func() {
			templateCmdData = oldData
			loader.GlobalLoadOptions = oldLoadOptions
		})
		templateCmdData = common.CmdData{}
		directory := ginkgo.GinkgoT().TempDir()
		var expected []string
		if bare {
			file := filepath.Join(directory, "bare.yaml")
			ginkgo.GinkgoT().Setenv("WERF_SECRET_VALUES", file)
			expected = append(expected, file)
		}
		if suffixed {
			file := filepath.Join(directory, "suffix.yaml")
			ginkgo.GinkgoT().Setenv("WERF_SECRET_VALUES_1", file)
			expected = append(expected, file)
		}
		cliFile := filepath.Join(directory, "cli.yaml")
		expected = append(expected, cliFile)
		ctx := context.Background()
		wc := chart_extender.NewWerfChartStub(ctx, false)
		loader.GlobalLoadOptions = &loader.LoadOptions{ChartExtender: wc}
		namespace := "default"
		cmd := NewTemplateCmd(&action.Configuration{}, wc, &namespace)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs([]string{"review", "testdata/chart", "--secret-values=" + cliFile})
		err := cmd.ExecuteContext(ctx)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unable to read custom secret values file \"" + expected[0] + "\"")))
		gomega.Expect(wc.SecretValueFiles).To(gomega.Equal(expected))
		gomega.Expect(InitRenderRelatedWerfChartParams(ctx, &templateCmdData, wc)).To(gomega.Succeed())
		gomega.Expect(wc.SecretValueFiles).To(gomega.Equal(expected))
	},
		ginkgo.Entry("bare environment", true, false),
		ginkgo.Entry("suffixed environment", false, true),
		ginkgo.Entry("bare then suffixed then flags", true, true),
		ginkgo.Entry("flags only", false, false),
	)
})
