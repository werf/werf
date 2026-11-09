package instruction

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/test/pkg/buildahstub"
)

type copyBackendStub struct {
	buildahstub.BuildahStub
	failure        string
	cleanupFailure bool
	err            error
	calls          []string
	cancel         context.CancelFunc
}

var _ buildah.Buildah = (*copyBackendStub)(nil)

func (backend *copyBackendStub) record(operation string) error {
	backend.calls = append(backend.calls, operation)
	if backend.failure == operation {
		backend.err = errors.New(operation + " failed")
		return backend.err
	}
	return nil
}

func (backend *copyBackendStub) FromCommand(_ context.Context, name, image string, _ buildah.FromCommandOpts) (string, error) {
	gomega.Expect(name).To(gomega.BeEmpty())
	gomega.Expect(image).To(gomega.Equal("source-image"))
	if err := backend.record("from"); err != nil {
		return "", err
	}
	if backend.failure == "save" {
		backend.err = errors.New("save failed")
		return "source-container", backend.err
	}
	return "source-container", nil
}

func (backend *copyBackendStub) Mount(_ context.Context, name string, _ buildah.MountOpts) (string, error) {
	gomega.Expect(name).To(gomega.Equal("source-container"))
	return "/source", backend.record("mount")
}

func (backend *copyBackendStub) Copy(ctx context.Context, name, dir string, _ []string, _ string, _ buildah.CopyOpts) error {
	gomega.Expect(name).To(gomega.Equal("destination"))
	gomega.Expect(dir).To(gomega.Equal("/source"))
	err := backend.record("copy")
	if backend.failure == "panic" {
		panic("copy panicked")
	}
	if backend.failure == "cancel" {
		backend.cancel()
		backend.err = ctx.Err()
		return backend.err
	}
	return err
}

func (backend *copyBackendStub) Umount(ctx context.Context, name string, _ buildah.UmountOpts) error {
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	gomega.Expect(name).To(gomega.Equal("source-container"))
	return backend.record("unmount")
}

func (backend *copyBackendStub) Rm(ctx context.Context, name string, _ buildah.RmOpts) error {
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	gomega.Expect(name).To(gomega.Equal("source-container"))
	err := backend.record("remove")
	if backend.cleanupFailure {
		return errors.New("remove failed")
	}
	return err
}

type copyArchiveStub struct{}

var _ container_backend.BuildContextArchiver = (*copyArchiveStub)(nil)

func (*copyArchiveStub) ExtractOrGetExtractedDir(context.Context) (string, error) {
	return "/source", nil
}

func (*copyArchiveStub) Create(context.Context, container_backend.BuildContextArchiveCreateOptions) error {
	panic("unexpected archive Create")
}

func (*copyArchiveStub) Path() string {
	panic("unexpected archive Path")
}

func (*copyArchiveStub) Open(context.Context) (io.ReadCloser, error) {
	panic("unexpected archive Open")
}

func (*copyArchiveStub) CalculatePathsChecksum(context.Context, []string) (string, error) {
	panic("unexpected archive CalculatePathsChecksum")
}

func (*copyArchiveStub) CalculateGlobsChecksum(context.Context, []string, container_backend.CalculateGlobsChecksumOptions) (string, error) {
	panic("unexpected archive CalculateGlobsChecksum")
}

func (*copyArchiveStub) CleanupExtractedDir(context.Context) {
	panic("unexpected archive CleanupExtractedDir")
}

type heredocCopyCall struct {
	Dir      string
	Sources  []string
	Dest     string
	Chown    string
	Chmod    string
	Contents map[string]string
}

type heredocBackendStub struct {
	buildahstub.BuildahStub
	copies       []heredocCopyCall
	adds         [][]string
	heredocDirs  []string
	heredocModes map[string]os.FileMode
}

var _ buildah.Buildah = (*heredocBackendStub)(nil)

func (backend *heredocBackendStub) Copy(_ context.Context, name, dir string, src []string, dst string, opts buildah.CopyOpts) error {
	gomega.Expect(name).To(gomega.Equal("destination"))
	call := heredocCopyCall{Dir: dir, Sources: src, Dest: dst, Chown: opts.Chown, Chmod: opts.Chmod}
	if dir != "/source" {
		call.Contents = make(map[string]string)
		for _, s := range src {
			data, err := os.ReadFile(filepath.Join(dir, s))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			call.Contents[s] = string(data)

			info, err := os.Stat(filepath.Join(dir, s))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			if backend.heredocModes == nil {
				backend.heredocModes = make(map[string]os.FileMode)
			}
			backend.heredocModes[s] = info.Mode().Perm()
		}
		call.Dir = ""
		backend.heredocDirs = append(backend.heredocDirs, dir)
	}
	backend.copies = append(backend.copies, call)
	return nil
}

func (backend *heredocBackendStub) Add(_ context.Context, name string, src []string, _ string, _ buildah.AddOpts) error {
	gomega.Expect(name).To(gomega.Equal("destination"))
	backend.adds = append(backend.adds, src)
	return nil
}
