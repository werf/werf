package filemanager_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"

	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader"
	filemanager "github.com/werf/werf/v2/pkg/giterminism_manager/file_manager"
)

var _ = ginkgo.Describe("LoadChartDir as the chart loader for a file:// dependency", func() {
	// The chart directory stays relative to the project directory all the way from werf to the
	// nelm loader, and the dependency path is resolved against it, so an absolute path here would
	// stop the includes from ever matching the dependent chart.
	const chartDir = ".helm"

	loadChartWithDependencies := func(ctx ginkgo.SpecContext, manager *filemanager.FileManager) []*nelmcommon.BufferedFile {
		parentFiles, err := manager.LoadChartDir(ctx, chartDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		files, err := loader.LoadChartDependencies(ctx, manager.LoadChartDir, chartDir, parentFiles, nelmcommon.HelmOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		return files
	}

	chartFileNames := func(files []*nelmcommon.BufferedFile) []string {
		return lo.Map(files, func(file *nelmcommon.BufferedFile, _ int) string { return file.Name })
	}

	ginkgo.It("filters the dependent chart by its own .helmignore and not by the parent chart rules", func(ctx ginkgo.SpecContext) {
		files := loadChartWithDependencies(ctx, newFileURLChartFileManager(ctx, false))

		gomega.Expect(chartFileNames(files)).To(gomega.ConsistOf(
			".helmignore",
			"Chart.yaml",
			"Chart.lock",
			"templates/root-kept.yaml",
			"charts/sub/.helmignore",
			"charts/sub/Chart.yaml",
			"charts/sub/templates/subchart-kept.yaml",
		))
	})

	ginkgo.It("adds the files included into the dependent chart and filters them by its own .helmignore", func(ctx ginkgo.SpecContext) {
		files := loadChartWithDependencies(ctx, newFileURLChartFileManager(ctx, true))

		gomega.Expect(chartFileNames(files)).To(gomega.ConsistOf(
			".helmignore",
			"Chart.yaml",
			"Chart.lock",
			"templates/root-kept.yaml",
			"charts/sub/.helmignore",
			"charts/sub/Chart.yaml",
			"charts/sub/templates/subchart-kept.yaml",
			"charts/sub/templates/from-include.yaml",
		))
		gomega.Expect(files).To(gomega.ContainElement(gomega.SatisfyAll(
			gomega.HaveField("Name", "charts/sub/templates/from-include.yaml"),
			gomega.HaveField("Data", []byte("subchart from include")),
		)))
	})
})
