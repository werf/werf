package migrate

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestMigrate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Meta-repo Migrate Suite")
}

var _ = ginkgo.Describe("migrate command", func() {
	ginkgo.It("scopes source and destination address env to the command", func() {
		ginkgo.GinkgoT().Setenv("WERF_FROM", "generic-from")
		ginkgo.GinkgoT().Setenv("WERF_TO", "generic-to")
		ginkgo.GinkgoT().Setenv("WERF_META_REPO_MIGRATE_FROM", "registry.example.com/project")
		ginkgo.GinkgoT().Setenv("WERF_META_REPO_MIGRATE_TO", "registry.example.com/project-meta")

		cmd := NewCmd(context.Background())

		fromFlag := cmd.Flags().Lookup("from")
		gomega.Expect(fromFlag).NotTo(gomega.BeNil())
		gomega.Expect(fromFlag.DefValue).To(gomega.Equal("registry.example.com/project"))

		toFlag := cmd.Flags().Lookup("to")
		gomega.Expect(toFlag).NotTo(gomega.BeNil())
		gomega.Expect(toFlag.DefValue).To(gomega.Equal("registry.example.com/project-meta"))

		gomega.Expect(cmd.Flags().Lookup("repo")).To(gomega.BeNil())
		gomega.Expect(cmd.Flags().Lookup("meta-repo")).To(gomega.BeNil())
	})

	ginkgo.It("exposes command-scoped source registry credential flags but not the destination", func() {
		ginkgo.GinkgoT().Setenv("WERF_META_REPO_MIGRATE_FROM_QUAY_TOKEN", "tok")

		cmd := NewCmd(context.Background())

		gomega.Expect(cmd.Flags().Lookup("from-docker-hub-token")).NotTo(gomega.BeNil())
		gomega.Expect(cmd.Flags().Lookup("from-container-registry")).NotTo(gomega.BeNil())

		quay := cmd.Flags().Lookup("from-quay-token")
		gomega.Expect(quay).NotTo(gomega.BeNil())
		gomega.Expect(quay.DefValue).To(gomega.Equal("tok"))

		gomega.Expect(cmd.Flags().Lookup("to-docker-hub-token")).To(gomega.BeNil())
		gomega.Expect(cmd.Flags().Lookup("to-quay-token")).To(gomega.BeNil())
	})

	ginkgo.It("defaults remove source to true", func() {
		ginkgo.GinkgoT().Setenv("WERF_REMOVE_SOURCE", "")

		cmd := NewCmd(context.Background())
		flag := cmd.Flags().Lookup("remove-source")
		gomega.Expect(flag).NotTo(gomega.BeNil())
		gomega.Expect(flag.DefValue).To(gomega.Equal("true"))
	})
	ginkgo.It("rejects unsafe concurrency and timeout values before initialization", func() {
		for arg, message := range map[string]string{"--parallel-tasks-limit=0": "--parallel-tasks-limit must be positive", "--timeout=-1s": "--timeout must not be negative"} {
			cmd := NewCmd(context.Background())
			cmd.SetArgs([]string{arg})
			gomega.Expect(cmd.Execute()).To(gomega.MatchError(message))
		}
	})

	ginkgo.It("reads the migration limits from environment", func() {
		ginkgo.GinkgoT().Setenv("WERF_PARALLEL_TASKS_LIMIT", "7")
		ginkgo.GinkgoT().Setenv("WERF_META_REPO_MIGRATE_TIMEOUT", "30m")
		cmd := NewCmd(context.Background())
		gomega.Expect(cmd.Flags().Lookup("parallel-tasks-limit").DefValue).To(gomega.Equal("7"))
		gomega.Expect(cmd.Flags().Lookup("timeout").DefValue).To(gomega.Equal("30m0s"))
	})
})
