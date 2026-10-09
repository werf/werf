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

	DescribeTable("collects orphaned project dirs across werf versions",
		func(prefix string, age time.Duration, dryRun bool) {
			past := time.Now().Add(-age)
			foreignPaths := []string{
				"foreign-tool-project-data-123",
				"werf-v2.78.2-docker-config-123",
				"werf-v2.78.2-context-123",
			}
			for _, name := range foreignPaths {
				path := filepath.Join(werf.GetTmpDir(), name)
				Expect(os.Mkdir(path, 0o755)).To(Succeed())
				Expect(os.Chtimes(path, past, past)).To(Succeed())
			}
			shouldRun, err := ShouldRunAutoGC()
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldRun).To(BeFalse())

			dir := filepath.Join(werf.GetTmpDir(), prefix+"123456789")
			Expect(os.Mkdir(dir, 0o755)).To(Succeed())
			payload := filepath.Join(dir, "data.txt")
			Expect(os.WriteFile(payload, []byte("payload"), 0o644)).To(Succeed())
			Expect(os.Chtimes(dir, past, past)).To(Succeed())

			shouldRun, err = ShouldRunAutoGC()
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldRun).To(Equal(age >= projectDirMaxAge))
			Expect(RunGC(GinkgoT().Context(), RunGCOptions{DryRun: dryRun})).To(Succeed())
			if age >= projectDirMaxAge && !dryRun {
				Expect(dir).NotTo(BeADirectory())
			} else {
				Expect(payload).To(BeARegularFile())
			}
			for _, name := range foreignPaths {
				Expect(filepath.Join(werf.GetTmpDir(), name)).To(BeADirectory())
			}
		},
		Entry("current version", projectDirPrefix, projectDirMaxAge+time.Hour, false),
		Entry("v2.77.2", "werf-v2.77.2-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("v2.78.2", "werf-v2.78.2-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("v2.79.2", "werf-v2.79.2-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("v3.4.0", "werf-v3.4.0-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("prerelease", "werf-v2.79.3-alpha.1-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("unversioned legacy dir", "werf-project-data-", projectDirMaxAge+time.Hour, false),
		Entry("fresh unversioned legacy dir", "werf-project-data-", projectDirMaxAge-time.Hour, false),
		Entry("unversioned legacy dir in dry run", "werf-project-data-", projectDirMaxAge+time.Hour, true),
		Entry("fresh dir from another version", "werf-v2.78.2-project-data-", projectDirMaxAge-time.Hour, false),
		Entry("dry run", "werf-v2.78.2-project-data-", projectDirMaxAge+time.Hour, true),
	)

	It("is collected by the tmp GC when orphaned by a killed process", func() {
		orphaned, err := CreateProjectDir(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())
		inUse, err := CreateProjectDir(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())
		// neither is registered: the process was killed before DelegateCleanup
		defer func() { Expect(DelegateCleanup(GinkgoT().Context())).To(Succeed()) }()

		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(orphaned, past, past)).To(Succeed())

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		Expect(orphaned).NotTo(BeADirectory())
		Expect(inUse).To(BeADirectory())
	})

	It("never removes what a symlink in the tmp dir points at", func() {
		target := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(target, "data.txt"), []byte("payload"), 0o644)).To(Succeed())

		link := filepath.Join(werf.GetTmpDir(), "werf-v2.78.2-project-data-123")
		Expect(os.Symlink(target, link)).To(Succeed())
		// os.Chtimes follows the link, so the age of the link itself is faked instead
		stubs.Stub(&timeSince, func(time.Time) time.Duration { return projectDirMaxAge + time.Hour })

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		Expect(target).To(BeADirectory())
		Expect(filepath.Join(target, "data.txt")).To(BeARegularFile())
		// the link itself is werf-prefixed and orphaned, so it is swept
		Expect(link).NotTo(BeAnExistingFile())
	})

	It("never descends into a filesystem mounted under the tmp dir", func() {
		mounted := filepath.Join(werf.GetTmpDir(), "werf-v2.78.2-project-data-123")
		Expect(os.MkdirAll(mounted, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(mounted, "data.txt"), []byte("payload"), 0o644)).To(Succeed())
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(mounted, past, past)).To(Succeed())
		stubs.Stub(&onSameDevice, func(_, entry os.FileInfo) bool { return entry.Name() != filepath.Base(mounted) })

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		Expect(filepath.Join(mounted, "data.txt")).To(BeARegularFile())
	})

	It("leaves foreign entries of the tmp dir alone", func() {
		foreign := filepath.Join(werf.GetTmpDir(), "foreign-tool-data")
		Expect(os.MkdirAll(foreign, 0o755)).To(Succeed())
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(foreign, past, past)).To(Succeed())

		Expect(RunGC(GinkgoT().Context(), RunGCOptions{})).To(Succeed())
		Expect(foreign).To(BeADirectory())
	})
})
