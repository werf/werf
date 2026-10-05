//go:build !windows

package tmp_manager

import (
	"os"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("device identity of tmp entries", func() {
	It("tells a directory on the tmp filesystem from one on another device", func() {
		tmp, err := os.Stat(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		dev, err := os.Stat("/dev")
		Expect(err).NotTo(HaveOccurred())
		Expect(tmp.Sys().(*syscall.Stat_t).Dev).NotTo(Equal(dev.Sys().(*syscall.Stat_t).Dev), "fixture needs /dev on another device than the temp dir")

		Expect(sameDevice(tmp, tmp)).To(BeTrue())
		Expect(sameDevice(tmp, dev)).To(BeFalse())
	})
})
