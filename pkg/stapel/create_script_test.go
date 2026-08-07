//go:build !windows

package stapel

import (
	"os"
	"path/filepath"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateScript", func() {
	It("creates an executable script under a umask that strips the executable bit", func() {
		previousUmask := syscall.Umask(0o001)
		defer syscall.Umask(previousUmask)

		scriptPath := filepath.Join(GinkgoT().TempDir(), "scripts", "script")
		Expect(CreateScript(scriptPath, []string{"echo hello"})).To(Succeed())

		info, err := os.Stat(scriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))
	})
})
