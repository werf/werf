package suite_init

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
)

func (data *SuiteData) SetupProjectCleanup() bool {
	return ginkgo.BeforeEach(func() {
		projectName := data.ProjectName
		data.CleanupRepositories = nil
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
			gomega.Expect(contback.CleanupProject(ctx, projectName, contback.CleanupProjectOptions{Repositories: data.CleanupRepositories})).To(gomega.Succeed())
		}, ginkgo.NodeTimeout(2*time.Minute))
	})
}
