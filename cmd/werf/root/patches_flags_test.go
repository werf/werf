package root

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	bundle_apply "github.com/werf/werf/v2/cmd/werf/bundle/apply"
	bundle_plan "github.com/werf/werf/v2/cmd/werf/bundle/plan"
	bundle_render "github.com/werf/werf/v2/cmd/werf/bundle/render"
	"github.com/werf/werf/v2/cmd/werf/converge"
	"github.com/werf/werf/v2/cmd/werf/dismiss"
	"github.com/werf/werf/v2/cmd/werf/lint"
	"github.com/werf/werf/v2/cmd/werf/plan"
	"github.com/werf/werf/v2/cmd/werf/render"
	"github.com/werf/werf/v2/cmd/werf/rollback"
)

func TestPatchesFlags(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Patches Flags Suite")
}

var _ = ginkgo.DescribeTable("patches flags of a command",
	func(newCmd func(context.Context) *cobra.Command, renders bool) {
		cmd := newCmd(context.Background())

		gomega.Expect(cmd.Flags().Lookup("no-default-patches")).NotTo(gomega.BeNil())

		patchesFlag := cmd.Flags().Lookup("patches")
		gomega.Expect(patchesFlag).NotTo(gomega.BeNil())
		gomega.Expect(patchesFlag.Usage).To(gomega.ContainSubstring("diff patches for drift detection"))

		if renders {
			gomega.Expect(patchesFlag.Usage).To(gomega.ContainSubstring("render patches for rendered resources"))
		} else {
			gomega.Expect(patchesFlag.Usage).NotTo(gomega.ContainSubstring("render patches"))
		}
	},
	ginkgo.Entry("render", render.NewCmd, true),
	ginkgo.Entry("bundle render", bundle_render.NewCmd, true),
	ginkgo.Entry("lint", lint.NewCmd, true),
	ginkgo.Entry("converge", converge.NewCmd, true),
	ginkgo.Entry("plan", plan.NewCmd, true),
	ginkgo.Entry("bundle apply", bundle_apply.NewCmd, true),
	ginkgo.Entry("bundle plan", bundle_plan.NewCmd, true),
	ginkgo.Entry("dismiss", dismiss.NewCmd, false),
	ginkgo.Entry("rollback", rollback.NewCmd, false),
)
