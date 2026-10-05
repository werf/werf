package tmp_manager

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = Describe("project dir", func() {
	var stubs *gostub.Stubs

	BeforeEach(func() {
		stubs = gostub.New()
		DeferCleanup(stubs.Reset)

		stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
		Expect(werf.Init("", "")).To(Succeed())
	})

	It("is collected by the tmp GC when orphaned by a killed process", func() {
		orphaned, err := CreateProjectDir(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())
		inUse, err := CreateProjectDir(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())
		// neither is registered: the process was killed before DelegateCleanup
		defer func() { Expect(DelegateCleanup(GinkgoT().Context())).To(Succeed()) }()

		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(orphaned, past, past)).To(Succeed())

		Expect(RunGC(GinkgoT().Context(), false)).To(Succeed())
		Expect(orphaned).NotTo(BeADirectory())
		Expect(inUse).To(BeADirectory())
	})

	It("leaves foreign entries of the tmp dir alone", func() {
		foreign := filepath.Join(werf.GetTmpDir(), "foreign-tool-data")
		Expect(os.MkdirAll(foreign, 0o755)).To(Succeed())
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(foreign, past, past)).To(Succeed())

		Expect(RunGC(GinkgoT().Context(), false)).To(Succeed())
		Expect(foreign).To(BeADirectory())
	})
})
