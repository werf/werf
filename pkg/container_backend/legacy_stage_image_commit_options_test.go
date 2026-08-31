package container_backend

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LegacyStageImageContainer commit options", func() {
	newContainer := func() *LegacyStageImageContainer {
		return newLegacyStageImageContainer(NewLegacyStageImage(nil, "repo:tag", nil, ""))
	}

	It("fails when the base image config has not been read before the run", func() {
		c := newContainer()

		_, err := c.prepareCommitOptions()

		Expect(err).To(MatchError(ContainSubstring("are not prepared")))
	})

	It("restores the base image config read before the run", func() {
		c := newContainer()
		c.inheritedCommitOptions = newLegacyStageContainerOptions()
		c.inheritedCommitOptions.Entrypoint = `["/bin/base-entrypoint"]`
		c.inheritedCommitOptions.User = "base-user"
		c.commitChangeOptions.AddUser("stage-user")

		opts, err := c.prepareCommitOptions()

		Expect(err).To(Succeed())
		Expect(opts.Entrypoint).To(Equal(`["/bin/base-entrypoint"]`))
		Expect(opts.User).To(Equal("stage-user"))
	})
})
