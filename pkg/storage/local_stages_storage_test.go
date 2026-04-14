package storage

import (
	"bytes"
	"context"
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

func (b *localMutationBackendStub) TagImageByName(ctx context.Context, img container_backend.LegacyImageInterface) error {
	if img.BuiltID() != "" {
		return b.Tag(ctx, img.BuiltID(), img.Name(), container_backend.TagOpts{})
	}
	return nil
}

var _ = Describe("LocalStagesStorage", func() {
	It("retags mutated local image by its new built id", func(ctx SpecContext) {
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))

		backend := &localMutationBackendStub{}
		storage := NewLocalStagesStorage(backend)
		stageImage := container_backend.NewLegacyStageImage(nil, "tmp-scratch-compare:stage", backend, "")
		stageImage.SetBuiltID("sha256:before")

		err := storage.MutateAndPushImage(logCtx, "tmp-scratch-compare:stage", "tmp-scratch-compare:stage", image.SpecConfig{Labels: map[string]string{"werf-stage-content-digest": "digest"}}, stageImage)
		Expect(err).NotTo(HaveOccurred())
		Expect(backend.savedImageName).To(Equal("tmp-scratch-compare:stage"))
		Expect(backend.loaded).To(BeTrue())
		Expect(stageImage.BuiltID()).To(Equal("sha256:mutated"))
		Expect(backend.tagCalls).To(Equal([][2]string{{"sha256:mutated", "tmp-scratch-compare:stage"}}))
	})
})

func newTinyDockerSaveTar() []byte {
	ref, err := name.ParseReference("example.com/test:latest")
	Expect(err).NotTo(HaveOccurred())

	var buf bytes.Buffer
	Expect(tarball.Write(ref, empty.Image, &buf)).To(Succeed())

	return buf.Bytes()
}
