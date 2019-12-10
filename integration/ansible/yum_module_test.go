// +build integration

package ansible_test

import (
	"github.com/flant/werf/integration/utils/werfexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stapel builder", func() {
	Context("when building stapel image based on centos 6 and 7", func() {
		It("successfully installs packages using yum module", func(done Done) {
			Expect(werfBuild("yum1", werfexec.CommandOptions{})).To(Succeed())
			close(done)
		}, 120)
	})

	Context("when building stapel image based on centos 8", func() {
		It("successfully installs packages using yum module", func(done Done) {
			Skip("FIXME https://github.com/flant/werf/issues/1983")
			Expect(werfBuild("yum2", werfexec.CommandOptions{})).To(Succeed())
			close(done)
		}, 120)
	})
})
