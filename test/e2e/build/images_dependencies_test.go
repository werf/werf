package e2e_build_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfBuild(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"build"}, extraArgs...)...)
}

var _ = Describe("Images dependencies", Label("e2e", "build", "extra"), func() {
	BeforeEach(func() {
		Expect(werf.Init(SuiteData.TmpDir, "")).To(Succeed())
	})

	When("dockerfile image uses COPY --from stage and external image", func() {
		It("should build staged dockerfile with COPY --from correctly", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(
				ctx,
				SuiteData.ProjectName,
				"_fixtures/images_dependencies/state1",
				"initial commit",
			)
			env := map[string]string{}
			if host := os.Getenv("WERF_TEST_BUILDKIT_HOST"); host != "" {
				env["WERF_BUILDKIT_HOST"] = host
			}
			Expect(
				werfBuild(
					ctx,
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					liveexec.ExecCommandOptions{
						Env: env,
					},
				),
			).To(Succeed())
		})
	})
})
