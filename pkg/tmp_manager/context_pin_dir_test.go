package tmp_manager

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("context pin dir", func() {
	It("lives next to the local cache and is collected by the tmp GC", func(ctx SpecContext) {
		stubs := gostub.New()
		defer stubs.Reset()
		stubs.Stub(&registrator, newGCRegistrator())

		stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
		Expect(werf.Init("", "")).To(Succeed())

		dir, err := CreateContextPinDir(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(dir).To(HavePrefix(werf.GetServiceDir() + string(os.PathSeparator)))
		Expect(DelegateCleanup(ctx)).To(Succeed())

		Expect(RunGC(ctx, RunGCOptions{})).To(Succeed())
		Expect(dir).NotTo(BeADirectory())
	})

	It("is collected by the tmp GC when orphaned by a killed process", func(ctx SpecContext) {
		stubs := gostub.New()
		defer stubs.Reset()
		stubs.Stub(&registrator, newGCRegistrator())

		stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
		Expect(werf.Init("", "")).To(Succeed())

		orphaned, err := CreateContextPinDir(ctx)
		Expect(err).NotTo(HaveOccurred())
		inUse, err := CreateContextPinDir(ctx)
		Expect(err).NotTo(HaveOccurred())
		// neither is registered: the process was killed before DelegateCleanup
		defer func() { Expect(DelegateCleanup(ctx)).To(Succeed()) }()

		past := time.Now().Add(-contextPinMaxAge - time.Hour)
		Expect(os.Chtimes(orphaned, past, past)).To(Succeed())

		Expect(RunGC(ctx, RunGCOptions{})).To(Succeed())
		Expect(orphaned).NotTo(BeADirectory())
		Expect(inUse).To(BeADirectory())
	})
})
