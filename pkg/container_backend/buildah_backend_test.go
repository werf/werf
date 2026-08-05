package container_backend

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/containers/storage/types"
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
