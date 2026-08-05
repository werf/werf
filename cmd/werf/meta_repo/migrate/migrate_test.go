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
	It("uses source and destination flags without environment defaults", func() {
		GinkgoT().Setenv("WERF_FROM", "registry.example.com/project")
		GinkgoT().Setenv("WERF_TO", "registry.example.com/project-meta")

		cmd := NewCmd(context.Background())

		fromFlag := cmd.Flags().Lookup("from")
		Expect(fromFlag).NotTo(BeNil())
		Expect(fromFlag.DefValue).To(BeEmpty())

		toFlag := cmd.Flags().Lookup("to")
		Expect(toFlag).NotTo(BeNil())
		Expect(toFlag.DefValue).To(BeEmpty())

		Expect(cmd.Flags().Lookup("repo")).To(BeNil())
		Expect(cmd.Flags().Lookup("meta-repo")).To(BeNil())
	})

	It("exposes source registry credential flags for deletes but not the destination", func() {
		cmd := NewCmd(context.Background())

		Expect(cmd.Flags().Lookup("from-docker-hub-token")).NotTo(BeNil())
		Expect(cmd.Flags().Lookup("from-quay-token")).NotTo(BeNil())
		Expect(cmd.Flags().Lookup("from-container-registry")).NotTo(BeNil())

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
