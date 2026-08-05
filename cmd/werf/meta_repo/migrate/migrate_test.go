package migrate

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMigrate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meta-repo Migrate Suite")
}

var _ = Describe("migrate command", func() {
	It("scopes source and destination address env to the command", func() {
		GinkgoT().Setenv("WERF_FROM", "generic-from")
		GinkgoT().Setenv("WERF_TO", "generic-to")
		GinkgoT().Setenv("WERF_META_REPO_MIGRATE_FROM", "registry.example.com/project")
		GinkgoT().Setenv("WERF_META_REPO_MIGRATE_TO", "registry.example.com/project-meta")

		cmd := NewCmd(context.Background())

		fromFlag := cmd.Flags().Lookup("from")
		Expect(fromFlag).NotTo(BeNil())
		Expect(fromFlag.DefValue).To(Equal("registry.example.com/project"))

		toFlag := cmd.Flags().Lookup("to")
		Expect(toFlag).NotTo(BeNil())
		Expect(toFlag.DefValue).To(Equal("registry.example.com/project-meta"))

		Expect(cmd.Flags().Lookup("repo")).To(BeNil())
		Expect(cmd.Flags().Lookup("meta-repo")).To(BeNil())
	})

	It("exposes command-scoped source registry credential flags but not the destination", func() {
		GinkgoT().Setenv("WERF_META_REPO_MIGRATE_FROM_QUAY_TOKEN", "tok")

		cmd := NewCmd(context.Background())

		Expect(cmd.Flags().Lookup("from-docker-hub-token")).NotTo(BeNil())
		Expect(cmd.Flags().Lookup("from-container-registry")).NotTo(BeNil())

		quay := cmd.Flags().Lookup("from-quay-token")
		Expect(quay).NotTo(BeNil())
		Expect(quay.DefValue).To(Equal("tok"))

		Expect(cmd.Flags().Lookup("to-docker-hub-token")).To(BeNil())
		Expect(cmd.Flags().Lookup("to-quay-token")).To(BeNil())
	})

	It("defaults remove source to true", func() {
		GinkgoT().Setenv("WERF_REMOVE_SOURCE", "")

		cmd := NewCmd(context.Background())
		flag := cmd.Flags().Lookup("remove-source")
		Expect(flag).NotTo(BeNil())
		Expect(flag.DefValue).To(Equal("true"))
	})
})
