package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

type localMutationBackendStub struct {
	container_backend.ContainerBackend

	savedImageName string
	loaded         bool
	tagCalls       [][2]string
}

func (b *localMutationBackendStub) SaveImageToStream(_ context.Context, imageName string) (io.ReadCloser, error) {
	b.savedImageName = imageName
	return io.NopCloser(bytes.NewReader(newTinyDockerSaveTar())), nil
}

func (b *localMutationBackendStub) LoadImageFromStream(_ context.Context, input io.Reader) (string, error) {
	_, err := io.ReadAll(input)
	Expect(err).NotTo(HaveOccurred())
	b.loaded = true
	return "sha256:mutated", nil
}

func (b *localMutationBackendStub) Tag(ctx context.Context, ref, newRef string, opts container_backend.TagOpts) error {
	b.tagCalls = append(b.tagCalls, [2]string{ref, newRef})
	return nil
}

var _ = Describe("LocalStagesStorage", func() {
	It("tags the mutated local image under the destination reference", func(ctx SpecContext) {
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))

		backend := &localMutationBackendStub{}
		storage := NewLocalStagesStorage(backend)
		stageImage := container_backend.NewLegacyStageImage(nil, "tmp-scratch-compare:stage", backend, "")

		err := storage.MutateAndPushImage(logCtx, "tmp-scratch-compare:stage", "tmp-scratch-compare:content-tag", image.SpecConfig{Labels: map[string]string{"werf-stage-content-digest": "digest"}}, stageImage)
		Expect(err).NotTo(HaveOccurred())
		Expect(backend.savedImageName).To(Equal("tmp-scratch-compare:stage"))
		Expect(backend.loaded).To(BeTrue())
		Expect(backend.tagCalls).To(Equal([][2]string{{"sha256:mutated", "tmp-scratch-compare:content-tag"}}))
	})

	It("uses the native mutator when the backend supports it, bypassing save/load", func(ctx SpecContext) {
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))

		backend := &nativeMutatorBackendStub{localMutationBackendStub: localMutationBackendStub{}}
		storage := NewLocalStagesStorage(backend)
		stageImage := container_backend.NewLegacyStageImage(nil, "tmp-scratch-compare:stage", backend, "linux/amd64")

		newConfig := image.SpecConfig{Labels: map[string]string{"werf-stage-content-digest": "digest"}}
		err := storage.MutateAndPushImage(logCtx, "tmp-scratch-compare:stage", "tmp-scratch-compare:content-tag", newConfig, stageImage)
		Expect(err).NotTo(HaveOccurred())

		Expect(backend.nativeCalls).To(Equal(1))
		Expect(backend.nativeSrc).To(Equal("tmp-scratch-compare:stage"))
		Expect(backend.nativeDest).To(Equal("tmp-scratch-compare:content-tag"))
		Expect(backend.nativeConfig).To(Equal(newConfig))
		Expect(backend.nativeTargetPlatform).To(Equal("linux/amd64"))

		Expect(backend.savedImageName).To(BeEmpty())
		Expect(backend.loaded).To(BeFalse())
		Expect(backend.tagCalls).To(BeEmpty())
	})

	It("falls back to save/load when the native mutator declines the config", func(ctx SpecContext) {
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))

		backend := &nativeMutatorBackendStub{nativeErr: container_backend.ErrNativeMutationUnsupported}
		storage := NewLocalStagesStorage(backend)
		stageImage := container_backend.NewLegacyStageImage(nil, "tmp-scratch-compare:stage", backend, "")

		err := storage.MutateAndPushImage(logCtx, "tmp-scratch-compare:stage", "tmp-scratch-compare:content-tag", image.SpecConfig{}, stageImage)
		Expect(err).NotTo(HaveOccurred())

		Expect(backend.nativeCalls).To(Equal(1))
		Expect(backend.savedImageName).To(Equal("tmp-scratch-compare:stage"))
		Expect(backend.loaded).To(BeTrue())
		Expect(backend.tagCalls).To(Equal([][2]string{{"sha256:mutated", "tmp-scratch-compare:content-tag"}}))
	})

	It("propagates a native mutator error other than unsupported", func(ctx SpecContext) {
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))

		backend := &nativeMutatorBackendStub{nativeErr: fmt.Errorf("boom")}
		storage := NewLocalStagesStorage(backend)
		stageImage := container_backend.NewLegacyStageImage(nil, "tmp-scratch-compare:stage", backend, "")

		err := storage.MutateAndPushImage(logCtx, "tmp-scratch-compare:stage", "tmp-scratch-compare:content-tag", image.SpecConfig{}, stageImage)
		Expect(err).To(MatchError(ContainSubstring("boom")))

		Expect(backend.nativeCalls).To(Equal(1))
		Expect(backend.savedImageName).To(BeEmpty())
		Expect(backend.loaded).To(BeFalse())
		Expect(backend.tagCalls).To(BeEmpty())
	})
})

type nativeMutatorBackendStub struct {
	localMutationBackendStub

	nativeCalls          int
	nativeSrc            string
	nativeDest           string
	nativeConfig         image.SpecConfig
	nativeTargetPlatform string
	nativeErr            error
}

func (b *nativeMutatorBackendStub) MutateAndPushImageNative(_ context.Context, src, dest string, newConfig image.SpecConfig, targetPlatform string) error {
	b.nativeCalls++
	b.nativeSrc = src
	b.nativeDest = dest
	b.nativeConfig = newConfig
	b.nativeTargetPlatform = targetPlatform
	return b.nativeErr
}

func newTinyDockerSaveTar() []byte {
	ref, err := name.ParseReference("example.com/test:latest")
	Expect(err).NotTo(HaveOccurred())

	var buf bytes.Buffer
	Expect(tarball.Write(ref, empty.Image, &buf)).To(Succeed())

	return buf.Bytes()
}
