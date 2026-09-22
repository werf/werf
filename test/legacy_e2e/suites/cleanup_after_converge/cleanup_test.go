package cleanup_with_k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("cleanup command", func() {
	const (
		artifactCacheVersion1             = "1"
		artifactCacheVersion2             = "2"
		artifactData1                     = "1"
		artifactData2                     = "2"
		expectedStageCountAfterFirstBuild = 5
	)

	setImageCredentialsEnv := func() {
		SuiteData.Stubs.SetEnv("WERF_SET_IMAGE_CREDENTIALS_REGISTRY", "imageCredentials.registry="+suite_init.TestRegistry())
	}

	setupProject := func(ctx context.Context) {
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
		setImageCredentialsEnv()
	}

	runCommand := func(ctx context.Context, args ...string) {
		utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, args...)
	}

	BeforeEach(func(ctx SpecContext) {
		setupProject(ctx)
	})

	AfterEach(func(ctx SpecContext) {
		runCommand(ctx, "dismiss", "--with-namespace")
		runCommand(ctx, "purge")
	})

	Context("kubernetes based policy", func() {
		BeforeEach(func(ctx SpecContext) {
			SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion1)
			SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData1)
			runCommand(ctx, "build")
		})

		It("should remove all stages", func(ctx SpecContext) {
			Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild))
			runCommand(ctx, "cleanup")
			Expect(StagesCount(ctx)).Should(Equal(0))
		})

		It("should keep all stages", func(ctx SpecContext) {
			runCommand(ctx, "converge", "--require-built-images")
			Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild))
			runCommand(ctx, "cleanup")
			Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild))
		})

		Context("artifact", func() {
			BeforeEach(func() {
				SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion2)
			})

			It("should delete replaced stages when imported content is unchanged", func(ctx SpecContext) {
				SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData1)
				runCommand(ctx, "converge")

				Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild + 3))

				runCommand(ctx, "cleanup")

				Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild))
			})

			It("should delete replaced stages when imported content changes", func(ctx SpecContext) {
				SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData2)
				runCommand(ctx, "converge")

				Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild + 3))

				runCommand(ctx, "cleanup")

				Expect(StagesCount(ctx)).Should(Equal(expectedStageCountAfterFirstBuild))
			})
		})
	})
})
