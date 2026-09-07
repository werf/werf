package container_backend

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/containers/storage/types"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/test/pkg/buildahstub"
)

var _ = Describe("BuildahBackend pulledImageIDs", func() {
	var backend *BuildahBackend

	BeforeEach(func() {
		backend = &BuildahBackend{}
	})

	DescribeTable("getPulledImageID",
		func(storeRef, storePlatform, storeID, queryRef, queryPlatform string, expectOK bool, expectedID string) {
			backend.storePulledImageID(storeRef, storePlatform, storeID)
			id, ok := backend.getPulledImageID(queryRef, queryPlatform)
			Expect(ok).To(Equal(expectOK))
			if expectOK {
				Expect(id).To(Equal(expectedID))
			}
		},
		Entry("same ref and platform", "alpine:latest", "linux/amd64", "sha256:aaa", "alpine:latest", "linux/amd64", true, "sha256:aaa"),
		Entry("different platform", "alpine:latest", "linux/arm64", "sha256:bbb", "alpine:latest", "linux/arm64", true, "sha256:bbb"),
		Entry("digest ref", "alpine@sha256:abc123", "linux/arm64", "sha256:ccc", "alpine@sha256:abc123", "linux/arm64", true, "sha256:ccc"),
		Entry("wrong ref", "alpine:latest", "linux/amd64", "sha256:aaa", "ubuntu:latest", "linux/amd64", false, ""),
		Entry("wrong platform", "alpine:latest", "linux/amd64", "sha256:aaa", "alpine:latest", "linux/arm64", false, ""),
	)

	It("overwrites imageID on repeated pull for same ref+platform", func() {
		backend.storePulledImageID("alpine:latest", "linux/amd64", "sha256:old")
		backend.storePulledImageID("alpine:latest", "linux/amd64", "sha256:new")

		id, ok := backend.getPulledImageID("alpine:latest", "linux/amd64")
		Expect(ok).To(BeTrue())
		Expect(id).To(Equal("sha256:new"))
	})

	It("Rmi removes entry from cache", func() {
		backend.storePulledImageID("alpine:latest", "linux/arm64", "sha256:aaa")

		backend.pulledImageIDs.Delete(pulledImageKey{"alpine:latest", "linux/arm64"})

		_, ok := backend.getPulledImageID("alpine:latest", "linux/arm64")
		Expect(ok).To(BeFalse())
	})

	It("Rmi does not remove entry for a different platform", func() {
		backend.storePulledImageID("alpine:latest", "linux/amd64", "sha256:amd64")
		backend.storePulledImageID("alpine:latest", "linux/arm64", "sha256:arm64")

		backend.pulledImageIDs.Delete(pulledImageKey{"alpine:latest", "linux/arm64"})

		id, ok := backend.getPulledImageID("alpine:latest", "linux/amd64")
		Expect(ok).To(BeTrue())
		Expect(id).To(Equal("sha256:amd64"))
	})
})

var _ = Describe("resolveContainerRootPath", func() {
	It("resolves an absolute symlink against the container root, not the host root", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "usr", "bin", "gotestsum"), []byte("bin"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		resolved, err := resolveContainerRootPath(rootMount, "/bin/gotestsum")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "usr", "bin", "gotestsum")))
	})

	It("resolves a relative symlink within the container root", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.Symlink("usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		resolved, err := resolveContainerRootPath(rootMount, "/bin/tool")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "usr", "bin", "tool")))
	})

	It("keeps a symlink escaping the container root scoped to it", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.Symlink("../../../etc", filepath.Join(rootMount, "escape"))).To(Succeed())

		resolved, err := resolveContainerRootPath(rootMount, "/escape/passwd")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "etc", "passwd")))
	})

	It("returns the joined path for a nonexistent destination", func() {
		rootMount := GinkgoT().TempDir()

		resolved, err := resolveContainerRootPath(rootMount, "/app/newfile")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "app", "newfile")))
	})
})

var _ = Describe("resolveContainerRootPathNoFollow", func() {
	It("resolves absolute symlink parents but keeps the final component unresolved", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())
		Expect(os.Symlink("gotestsum-real", filepath.Join(rootMount, "usr", "bin", "gotestsum"))).To(Succeed())

		resolved, err := resolveContainerRootPathNoFollow(rootMount, "/bin/gotestsum")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "usr", "bin", "gotestsum")))

		info, err := os.Lstat(resolved)
		Expect(err).ToNot(HaveOccurred())
		Expect(info.Mode() & os.ModeSymlink).ToNot(BeZero())
	})

	It("keeps a path that is itself a symlink pointing at its unresolved location", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		resolved, err := resolveContainerRootPathNoFollow(rootMount, "/bin")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "bin")))
	})

	It("normalizes a trailing slash before splitting off the final component", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		resolved, err := resolveContainerRootPathNoFollow(rootMount, "/bin/")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "bin")))
	})

	It("clamps a parent symlink escaping the container root", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.Symlink("../../../etc", filepath.Join(rootMount, "escape"))).To(Succeed())

		resolved, err := resolveContainerRootPathNoFollow(rootMount, "/escape/passwd")
		Expect(err).ToNot(HaveOccurred())
		Expect(resolved).To(Equal(filepath.Join(rootMount, "etc", "passwd")))
	})
})

var _ = Describe("calculateDependencyImportChecksum", func() {
	calc := func(rootMount string, spec DependencyImportSpec) string {
		fromPath, err := resolveContainerRootPathNoFollow(rootMount, spec.FromPath)
		Expect(err).ToNot(HaveOccurred())
		checksum, err := calculateDependencyImportChecksum(context.Background(), fromPath, spec)
		Expect(err).ToNot(HaveOccurred())
		return checksum
	}

	// entries: {content or symlink target, keyed path}, in sorted walk order.
	expectedChecksum := func(entries ...[2]string) string {
		hash := md5.New()
		for _, e := range entries {
			fmt.Fprintf(hash, "%x  %s\n", md5.Sum([]byte(e[0])), e[1])
		}
		return fmt.Sprintf("%x", hash.Sum(nil))
	}

	It("keys hash entries by the configured FromPath, not the mount location", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "bin", "gotestsum"), []byte("gotestsum-binary"), 0o755)).To(Succeed())

		checksum := calc(rootMount, DependencyImportSpec{FromPath: "/bin/gotestsum"})
		Expect(checksum).To(Equal(expectedChecksum([2]string{"gotestsum-binary", "/bin/gotestsum"})))
	})

	It("is identical between two different mount locations with the same content", func() {
		spec := DependencyImportSpec{FromPath: "/app"}

		var checksums []string
		for range 2 {
			rootMount := GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(rootMount, "app"), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(rootMount, "app", "server"), []byte("server-binary"), 0o755)).To(Succeed())
			checksums = append(checksums, calc(rootMount, spec))
		}

		Expect(checksums[0]).To(Equal(checksums[1]))
	})

	It("is identical between a plain directory and the same content behind an absolute parent symlink", func() {
		plainRoot := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(plainRoot, "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(plainRoot, "bin", "gotestsum"), []byte("gotestsum-binary"), 0o755)).To(Succeed())

		usrmergeRoot := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(usrmergeRoot, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(usrmergeRoot, "usr", "bin", "gotestsum"), []byte("gotestsum-binary"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(usrmergeRoot, "bin"))).To(Succeed())

		spec := DependencyImportSpec{FromPath: "/bin/gotestsum"}
		Expect(calc(usrmergeRoot, spec)).To(Equal(calc(plainRoot, spec)))
	})

	It("hashes a FromPath that is itself a symlink as a symlink, not its target tree", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "usr", "bin", "gotestsum"), []byte("gotestsum-binary"), 0o755)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		checksum := calc(rootMount, DependencyImportSpec{FromPath: "/bin"})
		Expect(checksum).To(Equal(expectedChecksum([2]string{"/usr/bin", "/bin"})))
	})

	It("honors include and exclude globs", func() {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "app"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "app", "keep.txt"), []byte("keep"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "app", "skip.log"), []byte("skip"), 0o644)).To(Succeed())

		checksum := calc(rootMount, DependencyImportSpec{FromPath: "/app", IncludePaths: []string{"*.txt"}})
		Expect(checksum).To(Equal(expectedChecksum([2]string{"keep", "/app/keep.txt"})))

		checksum = calc(rootMount, DependencyImportSpec{FromPath: "/app", ExcludePaths: []string{"*.log"}})
		Expect(checksum).To(Equal(expectedChecksum([2]string{"keep", "/app/keep.txt"})))
	})
})

var _ = Describe("BuildahBackend applyRemoveData", func() {
	It("keeps a symlinked parent dir when removing the last file under it", func(ctx SpecContext) {
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "usr", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "usr", "bin", "last-file"), []byte("data"), 0o644)).To(Succeed())
		Expect(os.Symlink("/usr/bin", filepath.Join(rootMount, "bin"))).To(Succeed())

		backend := &BuildahBackend{}
		Expect(backend.applyRemoveData(ctx, &containerDesc{RootMount: rootMount}, []RemoveDataSpec{{
			Type:           RemoveExactPathWithEmptyParentDirs,
			Paths:          []string{"/bin/last-file"},
			KeepParentDirs: []string{"/bin"},
		}})).To(Succeed())

		Expect(filepath.Join(rootMount, "usr", "bin", "last-file")).ToNot(BeAnExistingFile())
		Expect(filepath.Join(rootMount, "usr", "bin")).To(BeADirectory())
	})

	It("never prunes empty parent dirs above the container root", func(ctx SpecContext) {
		rootMount := filepath.Join(GinkgoT().TempDir(), "merged")
		Expect(os.MkdirAll(filepath.Join(rootMount, "app"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "app", "last-file"), []byte("data"), 0o644)).To(Succeed())

		backend := &BuildahBackend{}
		Expect(backend.applyRemoveData(ctx, &containerDesc{RootMount: rootMount}, []RemoveDataSpec{{
			Type:  RemoveExactPathWithEmptyParentDirs,
			Paths: []string{"/app/last-file"},
		}})).To(Succeed())

		Expect(filepath.Join(rootMount, "app")).ToNot(BeADirectory())
		Expect(rootMount).To(BeADirectory())
	})
})

var _ = Describe("platformMatches", func() {
	DescribeTable("validates platform",
		func(os, arch, variant, targetPlatform string, expected bool) {
			inspect := &thirdparty.BuilderInfo{
				OCIv1: v1.Image{Platform: v1.Platform{OS: os, Architecture: arch, Variant: variant}},
			}
			Expect(platformMatches(inspect, targetPlatform)).To(Equal(expected))
		},
		Entry("exact match linux/amd64", "linux", "amd64", "", "linux/amd64", true),
		Entry("exact match linux/arm64", "linux", "arm64", "", "linux/arm64", true),
		Entry("match with variant", "linux", "arm64", "v8", "linux/arm64/v8", true),
		Entry("os mismatch", "linux", "amd64", "", "windows/amd64", false),
		Entry("arch mismatch", "linux", "amd64", "", "linux/arm64", false),
		Entry("variant mismatch", "linux", "arm64", "v7", "linux/arm64/v8", false),
		Entry("no variant in target", "linux", "arm64", "v8", "linux/arm64", true),
		Entry("single-part platform passes", "linux", "amd64", "", "linux", true),
		// target specifies variant but image has no variant stored — treat as match
		// (OCI default: arm64 without explicit variant is equivalent to v8)
		Entry("target has variant, image variant empty", "linux", "arm64", "", "linux/arm64/v8", true),
	)
})

var _ = Describe("BuildahBackend createContainers", func() {
	It("re-pulls image by ref when cached imageID is missing locally", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.FromCommandFunc = func(_ context.Context, _, imageRef string, _ buildah.FromCommandOpts) (string, error) {
			switch imageRef {
			case "sha256:stale":
				return "", fmt.Errorf("locating image with name %q: %w", imageRef, types.ErrImageUnknown)
			case "sha256:fresh":
				return "container-id", nil
			default:
				return "", errors.New("unexpected image ref")
			}
		}
		fakeBuildah.PullFunc = func(_ context.Context, ref string, _ buildah.PullOpts) (string, error) {
			Expect(ref).To(Equal("registry.example.org/project/stage:tag"))
			return "sha256:fresh", nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})
		backend.storePulledImageID("registry.example.org/project/stage:tag", "linux/amd64", "sha256:stale")

		containers, err := backend.createContainers(context.Background(), []string{"registry.example.org/project/stage:tag"}, CommonOpts{TargetPlatform: "linux/amd64"})
		Expect(err).ToNot(HaveOccurred())
		Expect(containers).To(HaveLen(1))
		Expect(fakeBuildah.FromCommandImages).To(Equal([]string{"sha256:stale", "sha256:fresh"}))
		Expect(fakeBuildah.PullRefs).To(Equal([]string{"registry.example.org/project/stage:tag"}))

		cachedID, ok := backend.getPulledImageID("registry.example.org/project/stage:tag", "linux/amd64")
		Expect(ok).To(BeTrue())
		Expect(cachedID).To(Equal("sha256:fresh"))
	})

	It("serializes re-pulls after cached imageID misses", func() {
		const (
			imageRef     = "registry.example.org/project/stage:tag"
			platform     = "linux/amd64"
			staleImageID = "sha256:stale"
			freshImageID = "sha256:fresh"
		)

		firstPullStarted := make(chan struct{})
		releaseFirstPull := make(chan struct{})
		secondPullStarted := make(chan struct{}, 1)
		var pullCalls atomic.Int32
		var staleFromCalls atomic.Int32

		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.FromCommandFunc = func(_ context.Context, _, imageRef string, _ buildah.FromCommandOpts) (string, error) {
			switch imageRef {
			case staleImageID:
				staleFromCalls.Add(1)
				return "", fmt.Errorf("locating image with name %q: %w", imageRef, types.ErrImageUnknown)
			case freshImageID:
				return "container-id", nil
			default:
				return "", fmt.Errorf("unexpected image ref %q", imageRef)
			}
		}
		fakeBuildah.PullFunc = func(_ context.Context, ref string, _ buildah.PullOpts) (string, error) {
			if ref != imageRef {
				return "", fmt.Errorf("unexpected image ref %q", ref)
			}

			switch pullCalls.Add(1) {
			case 1:
				close(firstPullStarted)
				<-releaseFirstPull
			case 2:
				secondPullStarted <- struct{}{}
			default:
				return "", errors.New("unexpected pull")
			}

			return freshImageID, nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})
		backend.storePulledImageID(imageRef, platform, staleImageID)

		createContainers := func() <-chan error {
			done := make(chan error, 1)
			go func() {
				_, err := backend.createContainers(context.Background(), []string{imageRef}, CommonOpts{TargetPlatform: platform})
				done <- err
			}()
			return done
		}

		firstDone := createContainers()
		Eventually(firstPullStarted).Should(BeClosed())

		secondDone := createContainers()
		Eventually(staleFromCalls.Load).Should(Equal(int32(2)))
		Consistently(secondPullStarted).ShouldNot(Receive())

		close(releaseFirstPull)
		Eventually(firstDone).Should(Receive(Succeed()))
		Eventually(secondDone).Should(Receive(Succeed()))
		Expect(pullCalls.Load()).To(Equal(int32(2)))
	})
})

type stubMountsInstruction struct {
	mounts    []*instructions.Mount
	applyFunc func() error
}

func (i *stubMountsInstruction) Name() string { return "RUN" }

func (i *stubMountsInstruction) Apply(_ context.Context, _ string, _ buildah.Buildah, _ buildah.CommonOpts, _ BuildContextArchiver) error {
	if i.applyFunc != nil {
		return i.applyFunc()
	}
	return nil
}

func (i *stubMountsInstruction) UsesBuildContext() bool { return false }

func (i *stubMountsInstruction) GetMounts() []*instructions.Mount { return i.mounts }

var _ = Describe("BuildahBackend ensureRunMountImages", func() {
	const (
		imageRef = "registry.example.org/project/stage:tag"
		platform = "linux/amd64"
	)

	newRunInstruction := func(from string) *stubMountsInstruction {
		return &stubMountsInstruction{mounts: []*instructions.Mount{
			{Type: instructions.MountTypeBind, From: from, Target: "/mnt"},
			{Type: instructions.MountTypeTmpfs, Target: "/tmp"},
		}}
	}

	It("pulls the run mount from-image missing in local storage", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.PullFunc = func(_ context.Context, ref string, _ buildah.PullOpts) (string, error) {
			Expect(ref).To(Equal(imageRef))
			return "sha256:fresh", nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		err := backend.ensureRunMountImages(context.Background(), []InstructionInterface{newRunInstruction(imageRef)}, CommonOpts{TargetPlatform: platform})
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeBuildah.InspectRefs).To(Equal([]string{imageRef, imageRef}))
		Expect(fakeBuildah.PullRefs).To(Equal([]string{imageRef}))

		cachedID, ok := backend.getPulledImageID(imageRef, platform)
		Expect(ok).To(BeTrue())
		Expect(cachedID).To(Equal("sha256:fresh"))
	})

	It("does not pull when the from-image is already in local storage", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.InspectFunc = func(_ context.Context, _ string) (*thirdparty.BuilderInfo, error) {
			return &thirdparty.BuilderInfo{OCIv1: v1.Image{Platform: v1.Platform{OS: "linux", Architecture: "amd64"}}}, nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		err := backend.ensureRunMountImages(context.Background(), []InstructionInterface{newRunInstruction(imageRef)}, CommonOpts{TargetPlatform: platform})
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeBuildah.PullRefs).To(BeEmpty())
	})

	It("fails without pulling when the local from-image platform mismatches the target platform", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.InspectFunc = func(_ context.Context, _ string) (*thirdparty.BuilderInfo, error) {
			return &thirdparty.BuilderInfo{OCIv1: v1.Image{Platform: v1.Platform{OS: "linux", Architecture: "arm64"}}}, nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		err := backend.ensureRunMountImages(context.Background(), []InstructionInterface{newRunInstruction(imageRef)}, CommonOpts{TargetPlatform: platform})
		Expect(err).To(MatchError(ContainSubstring("but target platform is")))
		Expect(fakeBuildah.PullRefs).To(BeEmpty())
	})

	It("does not pull again when a concurrent caller already pulled the image", func() {
		var pulled atomic.Bool
		var inspectCalls atomic.Int32
		firstPullStarted := make(chan struct{})
		releaseFirstPull := make(chan struct{})
		var pullCalls atomic.Int32

		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.InspectFunc = func(_ context.Context, _ string) (*thirdparty.BuilderInfo, error) {
			inspectCalls.Add(1)
			if pulled.Load() {
				return &thirdparty.BuilderInfo{OCIv1: v1.Image{Platform: v1.Platform{OS: "linux", Architecture: "amd64"}}}, nil
			}
			return nil, nil
		}
		fakeBuildah.PullFunc = func(_ context.Context, ref string, _ buildah.PullOpts) (string, error) {
			if ref != imageRef {
				return "", fmt.Errorf("unexpected image ref %q", ref)
			}
			if pullCalls.Add(1) > 1 {
				return "", errors.New("unexpected second pull")
			}
			close(firstPullStarted)
			<-releaseFirstPull
			pulled.Store(true)
			return "sha256:fresh", nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		ensure := func() <-chan error {
			done := make(chan error, 1)
			go func() {
				done <- backend.ensureImageLocally(context.Background(), imageRef, CommonOpts{TargetPlatform: platform})
			}()
			return done
		}

		firstDone := ensure()
		Eventually(firstPullStarted).Should(BeClosed())

		secondDone := ensure()
		Eventually(inspectCalls.Load).Should(BeNumerically(">=", 3))

		close(releaseFirstPull)
		Eventually(firstDone).Should(Receive(Succeed()))
		Eventually(secondDone).Should(Receive(Succeed()))
		Expect(pullCalls.Load()).To(Equal(int32(1)))
	})

	It("fails when the pull fails", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.PullFunc = func(_ context.Context, _ string, _ buildah.PullOpts) (string, error) {
			return "", errors.New("no such host")
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		err := backend.ensureRunMountImages(context.Background(), []InstructionInterface{newRunInstruction(imageRef)}, CommonOpts{TargetPlatform: platform})
		Expect(err).To(MatchError(ContainSubstring("unable to pull image")))
	})

	It("skips instructions without mounts and mounts without from", func() {
		fakeBuildah := &buildahstub.BuildahStub{}
		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		instrs := []InstructionInterface{
			&stubMountsInstruction{mounts: []*instructions.Mount{{Type: instructions.MountTypeBind, Target: "/ctx"}}},
		}

		err := backend.ensureRunMountImages(context.Background(), instrs, CommonOpts{TargetPlatform: platform})
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeBuildah.InspectRefs).To(BeEmpty())
		Expect(fakeBuildah.PullRefs).To(BeEmpty())
	})

	It("BuildDockerfileStage pulls the missing run mount image before applying instructions", func() {
		var events []string

		fakeBuildah := &buildahstub.BuildahStub{}
		fakeBuildah.PullFunc = func(_ context.Context, ref string, _ buildah.PullOpts) (string, error) {
			Expect(ref).To(Equal(imageRef))
			events = append(events, "pull")
			return "sha256:fresh", nil
		}

		instr := newRunInstruction(imageRef)
		instr.applyFunc = func() error {
			events = append(events, "apply")
			return nil
		}

		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})

		_, err := backend.BuildDockerfileStage(context.Background(), "base-image:tag", BuildDockerfileStageOptions{
			CommonOpts: CommonOpts{TargetPlatform: platform},
		}, instr)
		Expect(err).ToNot(HaveOccurred())
		Expect(events).To(Equal([]string{"pull", "apply"}))
	})
})
