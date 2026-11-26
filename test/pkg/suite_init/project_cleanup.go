package suite_init

import (
	"os"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
)

func (data *SuiteData) SetContainerBackendMode(mode string) {
	data.Stubs.SetEnv("WERF_BUILDAH_MODE", mode)
	data.cleanupBackendModes = append(data.cleanupBackendModes, mode)
}

func (data *SuiteData) SetupProjectCleanup() bool {
	return ginkgo.BeforeEach(func() {
		projectName := data.ProjectName
		data.CleanupRepositories = nil
		data.cleanupBackendModes = nil
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
			modes := data.cleanupBackendModes
			if mode := os.Getenv("WERF_BUILDAH_MODE"); len(modes) == 0 && (mode == "" || mode == "docker") {
				modes = []string{"docker"}
			}
			gomega.Expect(contback.CleanupProject(ctx, projectName, contback.CleanupProjectOptions{
				Repositories: data.CleanupRepositories,
				BackendModes: modes,
			})).To(gomega.Succeed())
		}, ginkgo.NodeTimeout(2*time.Minute))
	})
}
