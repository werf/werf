package host_cleaning

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/moby/sys/mountinfo"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/host_cleaning/units"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/mock"
)

var _ = ginkgo.Describe("project tmp permission cleanup", func() {
	var (
		stubs   *gostub.Stubs
		backend *mock.MockContainerBackend
		project string
		past    time.Time
	)

	ginkgo.BeforeEach(func() {
		stubs = gostub.New()
		ginkgo.DeferCleanup(stubs.Reset)
		backend = mock.NewMockContainerBackend(gomock.NewController(ginkgo.GinkgoT()))
		stubs.SetEnv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		stubs.SetEnv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
		gomega.Expect(os.MkdirAll(werf.GetLocalCacheDir(), 0o700)).To(gomega.Succeed())
		project = filepath.Join(werf.GetTmpDir(), "werf-v2.1.0-project-data-old")
		gomega.Expect(os.MkdirAll(filepath.Join(project, "cache"), 0o700)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(filepath.Join(project, "cache", "payload"), []byte("cache"), 0o600)).To(gomega.Succeed())
		past = time.Now().Add(-48 * time.Hour).Truncate(time.Second)
		gomega.Expect(os.Chtimes(project, past, past)).To(gomega.Succeed())
	})

	ginkgo.It("removes writable contents without starting a backend", func(ctx ginkgo.SpecContext) {
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.Succeed())
		gomega.Expect(project).NotTo(gomega.BeAnExistingFile())
	})

	ginkgo.It("falls back after partial removal and limits the bind to the selected dir", func(ctx ginkgo.SpecContext) {
		outside := ginkgo.GinkgoT().TempDir()
		sentinel := filepath.Join(outside, "keep")
		gomega.Expect(os.WriteFile(sentinel, []byte("keep"), 0o600)).To(gomega.Succeed())
		gomega.Expect(os.Symlink(outside, filepath.Join(project, "link"))).To(gomega.Succeed())
		writable := filepath.Join(project, "writable")
		gomega.Expect(os.WriteFile(writable, []byte("remove"), 0o600)).To(gomega.Succeed())
		gomega.Expect(os.Chtimes(project, past, past)).To(gomega.Succeed())
		stubs.Stub(&projectTmpRemoveAll, func(path string) error {
			gomega.Expect(path).To(gomega.Equal(project))
			gomega.Expect(os.Remove(writable)).To(gomega.Succeed())
			return &os.PathError{Op: "unlinkat", Path: filepath.Join(path, "cache", "payload"), Err: fs.ErrPermission}
		})
		backend.EXPECT().RemoveHostDirs(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, mountDir string, dirs []string) error {
			gomega.Expect(mountDir).To(gomega.Equal(project))
			gomega.Expect(filepath.Dir(mountDir)).To(gomega.Equal(werf.GetTmpDir()))
			gomega.Expect(dirs).To(gomega.ConsistOf(filepath.Join(mountDir, "cache"), filepath.Join(mountDir, "link")))
			for _, dir := range dirs {
				if err := os.RemoveAll(dir); err != nil {
					return err
				}
			}
			return nil
		})
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.Succeed())
		gomega.Expect(sentinel).To(gomega.BeARegularFile())
		entries, err := os.ReadDir(werf.GetTmpDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.BeEmpty())
	})

	ginkgo.It("does not elevate a non-permission failure", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpRemoveAll, func(string) error { return syscall.EIO })
		gomega.Expect(errors.Is(removeProjectTmpDir(ctx, backend, project), syscall.EIO)).To(gomega.BeTrue())
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.It("does not elevate deletion of another user's directory", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpRemoveAll, func(string) error { return fs.ErrPermission })
		stubs.Stub(&projectTmpLstat, func(path string) (os.FileInfo, error) {
			info, err := os.Lstat(path)
			if err == nil {
				info.Sys().(*syscall.Stat_t).Uid = uint32(os.Geteuid()) + 1
			}
			return info, err
		})
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.MatchError(gomega.ContainSubstring("owner differs from current user")))
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.It("rejects replacement of the selected directory before backend removal", func(ctx ginkgo.SpecContext) {
		outside := ginkgo.GinkgoT().TempDir()
		sentinel := filepath.Join(outside, "keep")
		gomega.Expect(os.WriteFile(sentinel, []byte("keep"), 0o600)).To(gomega.Succeed())
		stubs.Stub(&projectTmpRemoveAll, func(path string) error {
			gomega.Expect(os.Rename(path, path+"-original")).To(gomega.Succeed())
			gomega.Expect(os.Symlink(outside, path)).To(gomega.Succeed())
			return fs.ErrPermission
		})
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.MatchError(gomega.ContainSubstring("changed before removal")))
		gomega.Expect(sentinel).To(gomega.BeARegularFile())
		gomega.Expect(filepath.Join(project+"-original", "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.DescribeTable("preserves mounted filesystems before any removal", func(suffix string, relative bool) {
		mountPath := project + suffix
		stubs.Stub(&projectTmpMounts, func(filter mountinfo.FilterFunc) ([]*mountinfo.Info, error) {
			mount := &mountinfo.Info{Mountpoint: mountPath}
			skip, _ := filter(mount)
			gomega.Expect(skip).To(gomega.BeFalse())
			return []*mountinfo.Info{mount}, nil
		})
		path := project
		if relative {
			cwd, err := os.Getwd()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			path, err = filepath.Rel(cwd, project)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}
		gomega.Expect(removeProjectTmpDir(context.Background(), backend, path)).To(gomega.MatchError(gomega.ContainSubstring("containing mountpoint")))
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	}, ginkgo.Entry("at the selected dir", "", false), ginkgo.Entry("inside the selected dir", "/cache", false), ginkgo.Entry("inside a relative path", "/cache", true))

	ginkgo.It("rejects a bind source with a writable non-sticky ancestor", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpRemoveAll, func(string) error { return fs.ErrPermission })
		gomega.Expect(os.Chmod(werf.GetTmpDir(), 0o777)).To(gomega.Succeed())
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.MatchError(gomega.ContainSubstring("writable without the sticky bit")))
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.It("rejects a bind source with an ancestor owned by another user", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpRemoveAll, func(string) error { return fs.ErrPermission })
		stubs.Stub(&projectTmpLstat, func(path string) (os.FileInfo, error) {
			info, err := os.Lstat(path)
			if err == nil && path == werf.GetTmpDir() {
				info.Sys().(*syscall.Stat_t).Uid = uint32(os.Geteuid()) + 1
			}
			return info, err
		})
		gomega.Expect(removeProjectTmpDir(ctx, backend, project)).To(gomega.MatchError(gomega.ContainSubstring("is not owned by root or the current user")))
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.It("fails closed when mount information is unavailable", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpMounts, func(mountinfo.FilterFunc) ([]*mountinfo.Info, error) { return nil, syscall.EIO })
		gomega.Expect(errors.Is(removeProjectTmpDir(ctx, backend, project), syscall.EIO)).To(gomega.BeTrue())
		gomega.Expect(filepath.Join(project, "cache", "payload")).To(gomega.BeARegularFile())
	})

	ginkgo.It("keeps a failed fallback discoverable and warns through host cleanup", func(ctx ginkgo.SpecContext) {
		stubs.Stub(&projectTmpRemoveAll, func(string) error { return fs.ErrPermission })
		failure := errors.New("backend unavailable")
		var selected string
		backend.EXPECT().RemoveHostDirs(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, mountDir string, _ []string) error {
			selected = mountDir
			return failure
		})
		var output bytes.Buffer
		logger := logboek.Context(ctx).NewSubLogger(&output, &output)
		logger.Streams().DisableLineWrapping()
		ctxWithLogger := logboek.NewContext(ctx, logger)
		gomega.Expect(RunHostCleanup(ctxWithLogger, backend, HostCleanupOptions{
			AllowedLocalCacheVolumeUsage: lo.Must(units.ParseUnitValue("100")),
		})).To(gomega.Succeed())
		gomega.Expect(output.String()).To(gomega.ContainSubstring("WARNING: unable to remove tmp data"))
		gomega.Expect(output.String()).To(gomega.ContainSubstring("backend unavailable"))
		gomega.Expect(output.String()).To(gomega.ContainSubstring(selected))
		gomega.Expect(filepath.Join(selected, "cache", "payload")).To(gomega.BeARegularFile())
		shouldRun, err := tmp_manager.ShouldRunAutoGC()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(shouldRun).To(gomega.BeTrue())
		backend.EXPECT().RemoveHostDirs(ctx, project, gomock.Any()).Return(failure).Times(20)
		for attempt := 0; attempt < 20; attempt++ {
			gomega.Expect(errors.Is(removeProjectTmpDir(ctx, backend, project), failure)).To(gomega.BeTrue())
			info, err := os.Stat(project)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.ModTime()).To(gomega.Equal(past))
		}
		backend.EXPECT().RemoveHostDirs(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dirs []string) error {
			for _, dir := range dirs {
				if err := os.RemoveAll(dir); err != nil {
					return err
				}
			}
			return nil
		})
		gomega.Expect(tmp_manager.RunGC(ctx, tmp_manager.RunGCOptions{
			RemoveProjectDir: func(ctx context.Context, path string) error { return removeProjectTmpDir(ctx, backend, path) },
		})).To(gomega.Succeed())
		entries, err := os.ReadDir(werf.GetTmpDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.BeEmpty())
	})
})
