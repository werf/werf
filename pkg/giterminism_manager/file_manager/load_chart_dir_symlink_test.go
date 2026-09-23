package filemanager_test

import (
	"runtime"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"

	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/werf/v2/pkg/logging"
)

// The commit walk resolves every symlink it does not skip, so these specs need a real repository:
// the mocked walks of the other suites never resolve anything and cannot fail here.
var _ = ginkgo.Describe("LoadChartDir over committed symlinks under enforced giterminism", func() {
	ginkgo.BeforeEach(func() {
		if runtime.GOOS == "windows" {
			ginkgo.Skip("symlinks in the repository are not exercised on windows")
		}
	})

	names := func(files []*nelmcommon.BufferedFile) []string {
		return lo.Map(files, func(f *nelmcommon.BufferedFile, _ int) string { return f.Name })
	}

	ginkgo.It("excludes a symlink loop matched by .helmignore without resolving it", func(ctx ginkgo.SpecContext) {
		manager := newCommittedChartFileManager(ctx, map[string]string{
			".helm/.helmignore":         "loop*\n",
			".helm/Chart.yaml":          "name: test",
			".helm/templates/kept.yaml": "kept",
		}, map[string]string{
			".helm/loop":  "loop2",
			".helm/loop2": "loop",
		})

		files, err := manager.LoadChartDir(logging.WithLogger(ctx), ".helm")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(names(files)).To(gomega.ConsistOf(".helmignore", "Chart.yaml", "templates/kept.yaml"))
	})

	// The walk descends into a symlinked directory by calling itself and has to hand the skip
	// predicate down, or the loop below the symlink is resolved although a rule excludes it.
	ginkgo.It("excludes a symlink loop below a symlinked chart directory without resolving it", func(ctx ginkgo.SpecContext) {
		manager := newCommittedChartFileManager(ctx, map[string]string{
			".helm/.helmignore": "templates/loop*\n",
			".helm/Chart.yaml":  "name: test",
			"shared/kept.yaml":  "kept",
		}, map[string]string{
			".helm/templates": "../shared",
			"shared/loop":     "loop2",
			"shared/loop2":    "loop",
		})

		files, err := manager.LoadChartDir(logging.WithLogger(ctx), ".helm")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(names(files)).To(gomega.ConsistOf(".helmignore", "Chart.yaml", "templates/kept.yaml"))
	})
})
