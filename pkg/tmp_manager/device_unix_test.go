//go:build !windows

package tmp_manager

import (
	"os"
	"syscall"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("device identity", func() {
	ginkgo.It("distinguishes device IDs from file metadata", func() {
		info, err := os.Stat(ginkgo.GinkgoT().TempDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		a := deviceFileInfo{FileInfo: info, stat: &syscall.Stat_t{Dev: 1}}
		b := deviceFileInfo{FileInfo: info, stat: &syscall.Stat_t{Dev: 2}}
		gomega.Expect(sameDevice(a, a)).To(gomega.BeTrue())
		gomega.Expect(sameDevice(a, b)).To(gomega.BeFalse())
	})
})
