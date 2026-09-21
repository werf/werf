package filemanager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestFileManager(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "File Manager Suite")
}

var _ = ginkgo.Describe("LoadChartDir", func() {
	ginkgo.DescribeTable("loads included chart files relative to the chart directory",
		func(ctx ginkgo.SpecContext, chartDir, destination string, localOverride bool) {
			manager, projectDir, includedFiles := newChartFileManager(ctx, destination, localOverride)

			files, err := manager.LoadChartDir(ctx, chartDir)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			expected := includedFiles
			if localOverride {
				expected["values.yaml"] = "source: local\n"
			}
			if destination == "/" {
				for _, name := range []string{"werf-includes.yaml", "werf-includes.lock"} {
					data, err := os.ReadFile(filepath.Join(projectDir, name))
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					expected[name] = string(data)
				}
			}

			actual := make(map[string]string)
			for _, chartFile := range files {
				gomega.Expect(actual).NotTo(gomega.HaveKey(chartFile.Name), "duplicate chart file")
				actual[chartFile.Name] = string(chartFile.Data)
			}
			gomega.Expect(actual).To(gomega.Equal(expected))
		},
		ginkgo.Entry("project root", ".", "/", false),
		ginkgo.Entry("project root with trailing slash", "./", "/", false),
		ginkgo.Entry("normalized project root", "chart/..", "/", false),
		ginkgo.Entry("local files override root includes", ".", "/", true),
		ginkgo.Entry("default chart directory", ".helm", "/.helm", false),
		ginkgo.Entry("chart subdirectory", "chart", "/chart", true),
		ginkgo.Entry("normalized chart subdirectory", "./chart/", "/chart", true),
	)
})
