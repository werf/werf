package meta_repo_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/meta_repo/detach"
	"github.com/werf/werf/v2/cmd/werf/meta_repo/migrate"
)

func TestMetaRepo(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Meta repository commands")
}

var _ = ginkgo.Describe("Kubernetes synchronization", func() {
	ginkgo.DescribeTable("reports kube configuration errors without panicking", func(newCmd func(context.Context) *cobra.Command, repoArgs []string) {
		dir := newProject()
		missingConfig := filepath.Join(dir, "missing-kubeconfig")
		common.SetupOndemandKubeInitializer("stale-context", "stale-config", "", []string{}, "", "")
		cmd := newCmd(context.Background())
		for _, flag := range []string{"final-repo", "secondary-repo", "cache-repo"} {
			gomega.Expect(cmd.Flags().Lookup(flag)).To(gomega.BeNil())
		}
		cmd.SetArgs(append(repoArgs, "--dir", dir, "--home-dir", filepath.Join(dir, "home"), "--tmp-dir", filepath.Join(dir, "tmp"), "--kube-config", missingConfig, "--synchronization", "kubernetes://meta-repo-test"))
		var err error
		gomega.Expect(func() { err = cmd.Execute() }).NotTo(gomega.Panic())
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(missingConfig)))
	},
		ginkgo.Entry("migrate", migrate.NewCmd, []string{"--from", "example.invalid/source", "--to", "example.invalid/meta"}),
		ginkgo.Entry("detach", detach.NewCmd, []string{"--repo", "example.invalid/source"}),
	)
})
