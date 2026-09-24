package tmp_manager

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = Describe("gc registration queue", func() {
	It("registers a queued path once", func() {
		stubs := gostub.New()
		defer stubs.Reset()

		for range 2 {
			stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
			stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
			Expect(werf.Init("", "")).To(Succeed())

			fromDockerConfig := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(fromDockerConfig, "config.json"), []byte("{}"), 0o600)).To(Succeed())

			_, err := CreateDockerConfigDir(GinkgoT().Context(), fromDockerConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(DelegateCleanup(GinkgoT().Context())).To(Succeed())
		}
	})
})
