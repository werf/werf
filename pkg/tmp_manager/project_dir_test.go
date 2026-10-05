package tmp_manager

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v2/pkg/werf"
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
		orphaned, err := CreateProjectDir(context.Background())
		Expect(err).NotTo(HaveOccurred())
		inUse, err := CreateProjectDir(context.Background())
		Expect(err).NotTo(HaveOccurred())
		// neither is registered: the process was killed before DelegateCleanup
		defer func() { Expect(DelegateCleanup(context.Background())).To(Succeed()) }()

		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(orphaned, past, past)).To(Succeed())

		Expect(RunGC(context.Background(), false)).To(Succeed())
		Expect(orphaned).NotTo(BeADirectory())
		Expect(inUse).To(BeADirectory())
	})

	It("never removes what a symlink in the tmp dir points at", func() {
		target := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(target, "data.txt"), []byte("payload"), 0o644)).To(Succeed())

		link := filepath.Join(werf.GetTmpDir(), projectDirPrefix+"foreign")
		Expect(os.Symlink(target, link)).To(Succeed())
		// os.Chtimes follows the link, so the age of the link itself is faked instead
		stubs.Stub(&timeSince, func(time.Time) time.Duration { return projectDirMaxAge + time.Hour })

		Expect(RunGC(context.Background(), false)).To(Succeed())
		Expect(target).To(BeADirectory())
		Expect(filepath.Join(target, "data.txt")).To(BeARegularFile())
		// the link itself is werf-prefixed and orphaned, so it is swept
		Expect(link).NotTo(BeAnExistingFile())
	})

	It("never descends into a filesystem mounted under the tmp dir", func() {
		mounted := filepath.Join(werf.GetTmpDir(), projectDirPrefix+"mounted")
		Expect(os.MkdirAll(mounted, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(mounted, "data.txt"), []byte("payload"), 0o644)).To(Succeed())
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(mounted, past, past)).To(Succeed())
		stubs.Stub(&onSameDevice, func(_, entry os.FileInfo) bool { return entry.Name() != filepath.Base(mounted) })

		Expect(RunGC(context.Background(), false)).To(Succeed())
		Expect(filepath.Join(mounted, "data.txt")).To(BeARegularFile())
	})

	It("leaves foreign entries of the tmp dir alone", func() {
		foreign := filepath.Join(werf.GetTmpDir(), "foreign-tool-data")
		Expect(os.MkdirAll(foreign, 0o755)).To(Succeed())
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		Expect(os.Chtimes(foreign, past, past)).To(Succeed())

		Expect(RunGC(context.Background(), false)).To(Succeed())
		Expect(foreign).To(BeADirectory())
	})
})
