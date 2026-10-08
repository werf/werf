package tmp_manager

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("project tmp removal callback", func() {
	ginkgo.BeforeEach(func() {
		stubs := gostub.New()
		ginkgo.DeferCleanup(stubs.Reset)
		stubs.SetEnv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("keeps the GC selection and dry-run contract", func(dryRun bool) {
		old := filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-project-data-old")
		legacy := filepath.Join(werf.GetTmpDir(), "werf-project-data-old")
		fresh := filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-project-data-fresh")
		foreign := filepath.Join(werf.GetTmpDir(), "foreign-project-data-old")
		past := time.Now().Add(-projectDirMaxAge - time.Hour)
		for _, path := range []string{old, legacy, fresh, foreign} {
			gomega.Expect(os.Mkdir(path, 0o700)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(path, "data"), []byte("payload"), 0o600)).To(gomega.Succeed())
			if path != fresh {
				gomega.Expect(os.Chtimes(path, past, past)).To(gomega.Succeed())
			}
		}
		var removed []string
		gomega.Expect(RunGC(context.Background(), RunGCOptions{
			DryRun: dryRun,
			RemoveProjectDir: func(ctx context.Context, path string) error {
				removed = append(removed, path)
				return os.RemoveAll(path)
			},
		})).To(gomega.Succeed())
		if dryRun {
			gomega.Expect(removed).To(gomega.BeEmpty())
			gomega.Expect(filepath.Join(old, "data")).To(gomega.BeARegularFile())
			gomega.Expect(filepath.Join(legacy, "data")).To(gomega.BeARegularFile())
		} else {
			gomega.Expect(removed).To(gomega.ConsistOf(old, legacy))
			gomega.Expect(old).NotTo(gomega.BeAnExistingFile())
			gomega.Expect(legacy).NotTo(gomega.BeAnExistingFile())
		}
		gomega.Expect(filepath.Join(fresh, "data")).To(gomega.BeARegularFile())
		gomega.Expect(filepath.Join(foreign, "data")).To(gomega.BeARegularFile())
	}, ginkgo.Entry("cleanup", false), ginkgo.Entry("dry run", true))

	ginkgo.It("uses the callback for registered projects from a previous tmp root", func(ctx ginkgo.SpecContext) {
		ordinary := filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-context-old")
		outside := filepath.Join(ginkgo.GinkgoT().TempDir(), "werf-v2.1.0-project-data-old")
		for _, path := range []string{ordinary, outside} {
			gomega.Expect(os.Mkdir(path, 0o700)).To(gomega.Succeed())
			gomega.Expect(registerPath(path, filepath.Join(getReleasedTmpDirs(), projectsServiceDir))).To(gomega.Succeed())
		}
		var removed []string
		gomega.Expect(RunGC(ctx, RunGCOptions{
			RemoveProjectDir: func(_ context.Context, path string) error {
				removed = append(removed, path)
				return os.RemoveAll(path)
			},
		})).To(gomega.Succeed())
		gomega.Expect(removed).To(gomega.ConsistOf(outside))
		gomega.Expect(ordinary).NotTo(gomega.BeAnExistingFile())
		gomega.Expect(outside).NotTo(gomega.BeAnExistingFile())
	})

	ginkgo.It("reports removal failures and continues cleaning other tmp data", func(ctx ginkgo.SpecContext) {
		project := filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-project-data-old")
		ordinary := filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-context-old")
		for _, path := range []string{project, ordinary} {
			gomega.Expect(os.Mkdir(path, 0o700)).To(gomega.Succeed())
			gomega.Expect(registerPath(path, filepath.Join(getReleasedTmpDirs(), projectsServiceDir))).To(gomega.Succeed())
		}
		failure := errors.New("backend unavailable")
		err := RunGC(ctx, RunGCOptions{
			RemoveProjectDir: func(context.Context, string) error { return failure },
		})
		gomega.Expect(errors.Is(err, failure)).To(gomega.BeTrue())
		gomega.Expect(errors.Is(err, ErrPathRemoval)).To(gomega.BeTrue())
		gomega.Expect(project).To(gomega.BeADirectory())
		gomega.Expect(ordinary).NotTo(gomega.BeAnExistingFile())
		links, err := os.ReadDir(filepath.Join(getReleasedTmpDirs(), projectsServiceDir))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(links).To(gomega.HaveLen(1))
		for round := 0; round < 3; round++ {
			config, err := os.MkdirTemp(werf.GetTmpDir(), "config-")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			registry := filepath.Join(getCreatedTmpDirs(), kubeConfigsServiceDir)
			gomega.Expect(registerPath(config, registry)).To(gomega.Succeed())
			err = RunGC(ctx, RunGCOptions{RemoveProjectDir: func(context.Context, string) error { return failure }})
			gomega.Expect(errors.Is(err, failure)).To(gomega.BeTrue())
			gomega.Expect(config).NotTo(gomega.BeAnExistingFile())
			remaining, err := os.ReadDir(registry)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(remaining).To(gomega.BeEmpty())
		}
		shouldRun, err := ShouldRunAutoGC()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(shouldRun).To(gomega.BeTrue())
		var retried []string
		gomega.Expect(RunGC(ctx, RunGCOptions{
			RemoveProjectDir: func(_ context.Context, path string) error {
				retried = append(retried, path)
				return os.RemoveAll(path)
			},
		})).To(gomega.Succeed())
		gomega.Expect(retried).To(gomega.ConsistOf(project))
	})
})
