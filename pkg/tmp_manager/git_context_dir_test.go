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

var _ = Describe("git context dir", func() {
	var stubs *gostub.Stubs

	BeforeEach(func() {
		stubs = gostub.New()
		DeferCleanup(stubs.Reset)

		stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
		Expect(werf.Init("", "")).To(Succeed())
	})

	DescribeTable("is collected by the tmp GC when orphaned by a killed process",
		func(age time.Duration, shouldBeRemoved bool) {
			dir, err := TempDir("git-context-")
			Expect(err).NotTo(HaveOccurred())
			payload := filepath.Join(dir, "data.txt")
			Expect(os.WriteFile(payload, []byte("payload"), 0o644)).To(Succeed())
			past := time.Now().Add(-age)
			Expect(os.Chtimes(dir, past, past)).To(Succeed())

			Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
			if shouldBeRemoved {
				Expect(dir).NotTo(BeADirectory())
			} else {
				Expect(payload).To(BeARegularFile())
			}
		},
		Entry("orphaned", gitContextDirMaxAge+time.Hour, true),
		Entry("fresh", gitContextDirMaxAge-time.Hour, false),
	)

	It("never removes what a symlink in the tmp dir points at", func() {
		target := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(target, "data.txt"), []byte("payload"), 0o644)).To(Succeed())

		link := filepath.Join(werf.GetTmpDir(), commonPrefix+"git-context-123")
		Expect(os.Symlink(target, link)).To(Succeed())
		// os.Chtimes follows the link, so the age of the link itself is faked instead
		stubs.Stub(&timeSince, func(time.Time) time.Duration { return gitContextDirMaxAge + time.Hour })

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		Expect(filepath.Join(target, "data.txt")).To(BeARegularFile())
		Expect(link).NotTo(BeAnExistingFile())
	})

	It("leaves foreign entries of the tmp dir alone", func() {
		past := time.Now().Add(-gitContextDirMaxAge - time.Hour)
		foreignPaths := []string{
			"git-context-123",
			commonPrefix + "git-contexts",
		}
		for _, name := range foreignPaths {
			path := filepath.Join(werf.GetTmpDir(), name)
			Expect(os.Mkdir(path, 0o755)).To(Succeed())
			Expect(os.Chtimes(path, past, past)).To(Succeed())
		}

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		for _, name := range foreignPaths {
			Expect(filepath.Join(werf.GetTmpDir(), name)).To(BeADirectory())
		}
	})
})
