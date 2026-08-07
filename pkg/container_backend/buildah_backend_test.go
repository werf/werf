package container_backend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.podman.io/storage"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/image"
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

type mutateNativeBuildahStub struct {
	fromCommandCalls     int
	fromCommandImage     string
	fromCommandContainer string
	rmCalls              []string

	mutateConfigCalls     int
	mutateConfigContainer string
	mutateConfigConfig    image.SpecConfig
	mutateConfigOpts      buildah.CommonOpts

	commitMutationCalls     int
	commitMutationContainer string
	commitMutationOpts      buildah.CommitOpts

	fromCommandErr    error
	mutateConfigErr   error
	commitMutationErr error
}

func (s *mutateNativeBuildahStub) FromCommand(_ context.Context, container, image string, _ buildah.FromCommandOpts) (string, error) {
	s.fromCommandCalls++
	s.fromCommandImage = image
	s.fromCommandContainer = container
	if s.fromCommandErr != nil {
		return "", s.fromCommandErr
	}
	return container, nil
}

func (s *mutateNativeBuildahStub) Rm(_ context.Context, ref string, _ buildah.RmOpts) error {
	s.rmCalls = append(s.rmCalls, ref)
	return nil
}

func (s *mutateNativeBuildahStub) MutateConfig(_ context.Context, container string, newConfig image.SpecConfig, opts buildah.CommonOpts) error {
	s.mutateConfigCalls++
	s.mutateConfigContainer = container
	s.mutateConfigConfig = newConfig
	s.mutateConfigOpts = opts
	return s.mutateConfigErr
}

func (s *mutateNativeBuildahStub) CommitMutation(_ context.Context, container string, opts buildah.CommitOpts) (string, error) {
	s.commitMutationCalls++
	s.commitMutationContainer = container
	s.commitMutationOpts = opts
	if s.commitMutationErr != nil {
		return "", s.commitMutationErr
	}
	return "new-image-id", nil
}

func (s *mutateNativeBuildahStub) Info(context.Context) (info.Info, error) { panic("unexpected call") }
func (s *mutateNativeBuildahStub) GetDefaultPlatform() string              { panic("unexpected call") }
func (s *mutateNativeBuildahStub) GetRuntimePlatform() string              { panic("unexpected call") }

func (s *mutateNativeBuildahStub) Tag(context.Context, string, string, buildah.TagOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Push(context.Context, string, buildah.PushOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) BuildFromDockerfile(context.Context, string, buildah.BuildFromDockerfileOpts) (string, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) RunCommand(context.Context, string, []string, buildah.RunCommandOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Pull(context.Context, string, buildah.PullOpts) (string, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Inspect(context.Context, string) (*thirdparty.BuilderInfo, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Rmi(context.Context, string, buildah.RmiOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Mount(context.Context, string, buildah.MountOpts) (string, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Umount(context.Context, string, buildah.UmountOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Commit(context.Context, string, buildah.CommitOpts) (string, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Config(context.Context, string, buildah.ConfigOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Copy(context.Context, string, string, []string, string, buildah.CopyOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Add(context.Context, string, []string, string, buildah.AddOpts) error {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Images(context.Context, buildah.ImagesOptions) (image.ImagesList, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) Containers(context.Context, buildah.ContainersOptions) (image.ContainerList, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) PruneImages(context.Context, buildah.PruneImagesOptions) (buildah.PruneImagesReport, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) SaveImageToStream(context.Context, string) (io.ReadCloser, error) {
	panic("unexpected call")
}

func (s *mutateNativeBuildahStub) LoadImageFromStream(context.Context, io.Reader) (string, error) {
	panic("unexpected call")
}

var _ = Describe("BuildahBackend.MutateAndPushImageNative", func() {
	It("creates a container from src, mutates its config, commits to dest, and removes the temp container", func(ctx SpecContext) {
		stub := &mutateNativeBuildahStub{}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})

		newConfig := image.SpecConfig{Labels: map[string]string{"foo": "bar"}}
		err := backend.MutateAndPushImageNative(ctx, "src:latest", "dest:latest", newConfig, "linux/amd64")
		Expect(err).NotTo(HaveOccurred())

		Expect(stub.fromCommandCalls).To(Equal(1))
		Expect(stub.fromCommandImage).To(Equal("src:latest"))

		Expect(stub.mutateConfigCalls).To(Equal(1))
		Expect(stub.mutateConfigConfig).To(Equal(newConfig))
		Expect(stub.mutateConfigOpts.TargetPlatform).To(Equal("linux/amd64"))
		Expect(stub.mutateConfigContainer).To(Equal(stub.fromCommandContainer))

		Expect(stub.commitMutationCalls).To(Equal(1))
		Expect(stub.commitMutationOpts.Image).To(Equal("dest:latest"))
		Expect(stub.commitMutationOpts.ClearHistory).To(BeFalse())
		Expect(stub.commitMutationOpts.Created).To(BeNil())
		Expect(stub.commitMutationContainer).To(Equal(stub.fromCommandContainer))

		Expect(stub.rmCalls).To(Equal([]string{stub.fromCommandContainer}))
	})

	It("propagates ClearHistory and parses Created into CommitMutation options", func(ctx SpecContext) {
		stub := &mutateNativeBuildahStub{}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})

		newConfig := image.SpecConfig{ClearHistory: true, Created: "2024-01-02T03:04:05Z"}
		err := backend.MutateAndPushImageNative(ctx, "src:latest", "dest:latest", newConfig, "")
		Expect(err).NotTo(HaveOccurred())

		Expect(stub.commitMutationOpts.ClearHistory).To(BeTrue())
		Expect(stub.commitMutationOpts.Created).NotTo(BeNil())
		Expect(stub.commitMutationOpts.Created.UTC().Format("2006-01-02T15:04:05Z")).To(Equal("2024-01-02T03:04:05Z"))
	})

	It("removes the temp container even when MutateConfig fails", func(ctx SpecContext) {
		stub := &mutateNativeBuildahStub{mutateConfigErr: fmt.Errorf("mutate boom")}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})

		err := backend.MutateAndPushImageNative(ctx, "src:latest", "dest:latest", image.SpecConfig{}, "")
		Expect(err).To(MatchError(ContainSubstring("mutate boom")))
		Expect(stub.commitMutationCalls).To(Equal(0))
		Expect(stub.rmCalls).To(Equal([]string{stub.fromCommandContainer}))
	})

	It("returns an error for an unparseable Created timestamp", func(ctx SpecContext) {
		stub := &mutateNativeBuildahStub{}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})

		err := backend.MutateAndPushImageNative(ctx, "src:latest", "dest:latest", image.SpecConfig{Created: "not-a-timestamp"}, "")
		Expect(err).To(HaveOccurred())
		Expect(stub.commitMutationCalls).To(Equal(0))
		Expect(stub.rmCalls).To(Equal([]string{stub.fromCommandContainer}))
	})

	It("propagates FromCommand errors without calling MutateConfig or CommitMutation", func(ctx SpecContext) {
		stub := &mutateNativeBuildahStub{fromCommandErr: fmt.Errorf("from boom")}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})

		err := backend.MutateAndPushImageNative(ctx, "src:latest", "dest:latest", image.SpecConfig{}, "")
		Expect(err).To(MatchError(ContainSubstring("from boom")))
		Expect(stub.mutateConfigCalls).To(Equal(0))
		Expect(stub.commitMutationCalls).To(Equal(0))
		Expect(stub.rmCalls).To(BeEmpty())
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
				return "", fmt.Errorf("locating image with name %q: %w", imageRef, storage.ErrImageUnknown)
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
				return "", fmt.Errorf("locating image with name %q: %w", imageRef, storage.ErrImageUnknown)
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
