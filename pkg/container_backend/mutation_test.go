package container_backend

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	werfimage "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/werf"
)

type fakeBackendLoaderStorer struct {
	saveStream io.ReadCloser
	saveErr    error
	loadErr    error

	loadCalls  int
	loadedData []byte
	loadedImg  v1.Image
}

func (b *fakeBackendLoaderStorer) SaveImageToStream(_ context.Context, _ string) (io.ReadCloser, error) {
	if b.saveErr != nil {
		return nil, b.saveErr
	}

	return b.saveStream, nil
}

func (b *fakeBackendLoaderStorer) LoadImageFromStream(_ context.Context, input io.Reader) (string, error) {
	b.loadCalls++

	data, err := io.ReadAll(input)
	if err != nil {
		return "", err
	}
	b.loadedData = data

	if b.loadErr != nil {
		return "", b.loadErr
	}

	img, err := tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}, nil)
	if err != nil {
		return "", err
	}
	b.loadedImg = img

	return "sha256:mutated", nil
}

type errReadCloser struct {
	err    error
	closed bool
}

func (r *errReadCloser) Read(_ []byte) (int, error) {
	return 0, r.err
}

func (r *errReadCloser) Close() error {
	r.closed = true
	return nil
}

var _ = Describe("MutateAndPushImage", func() {
	BeforeEach(func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())
	})

	It("persists docker save output to a temp tar file before mutating and loading it back", func(ctx SpecContext) {
		sourceImage := newSourceImage()
		sourceTarData := newDockerSaveTar(sourceImage)
		backend := &fakeBackendLoaderStorer{saveStream: io.NopCloser(bytes.NewReader(sourceTarData))}

		newID, err := MutateAndPushImage(ctx, "example.com/test:latest", "", werfimage.SpecConfig{
			Author:     "werf",
			Cmd:        []string{"/app", "serve"},
			Env:        []string{"KEY=VALUE"},
			Labels:     map[string]string{"mutated": "true"},
			WorkingDir: "/work",
		}, backend)
		Expect(err).ToNot(HaveOccurred())
		Expect(newID).To(Equal("sha256:mutated"))
		Expect(backend.loadCalls).To(Equal(1))
		Expect(backend.loadedData).ToNot(BeEmpty())
		Expect(backend.loadedImg).ToNot(BeNil())

		cfg, err := backend.loadedImg.ConfigFile()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Author).To(Equal("werf"))
		Expect(cfg.Config.Cmd).To(Equal([]string{"/app", "serve"}))
		Expect(cfg.Config.Env).To(Equal([]string{"KEY=VALUE"}))
		Expect(cfg.Config.WorkingDir).To(Equal("/work"))
		Expect(cfg.Config.Labels).To(Equal(map[string]string{"mutated": "true"}))
	})

	It("returns save errors before attempting to load the mutated image", func(ctx SpecContext) {
		backend := &fakeBackendLoaderStorer{saveErr: fmt.Errorf("save boom")}

		_, err := MutateAndPushImage(ctx, "example.com/test:latest", "", werfimage.SpecConfig{}, backend)
		Expect(err).To(MatchError("failed to save image: save boom"))
		Expect(backend.loadCalls).To(Equal(0))
	})

	It("returns tar persistence errors before attempting to load the mutated image", func(ctx SpecContext) {
		brokenStream := &errReadCloser{err: fmt.Errorf("stream boom")}
		backend := &fakeBackendLoaderStorer{saveStream: brokenStream}

		_, err := MutateAndPushImage(ctx, "example.com/test:latest", "", werfimage.SpecConfig{}, backend)
		Expect(err).To(MatchError("failed to persist image tarball: stream boom"))
		Expect(backend.loadCalls).To(Equal(0))
		Expect(brokenStream.closed).To(BeTrue())
	})

	It("preserves target platform in mutated image config", func(ctx SpecContext) {
		t := GinkgoT()

		sourceImage := newPlatformSourceImage()
		sourceTarData := newDockerSaveTar(sourceImage)
		backend := &fakeBackendLoaderStorer{saveStream: io.NopCloser(bytes.NewReader(sourceTarData))}

		newID, err := MutateAndPushImage(ctx, "example.com/test:latest", "linux/amd64", werfimage.SpecConfig{}, backend)
		require.NoError(t, err)
		assert.Equal(t, "sha256:mutated", newID)
		require.NotNil(t, backend.loadedImg)

		cfg, err := backend.loadedImg.ConfigFile()
		require.NoError(t, err)
		assert.Equal(t, "linux", cfg.OS)
		assert.Equal(t, "amd64", cfg.Architecture)
		assert.Empty(t, cfg.Variant)
	})
})

func newSourceImage() v1.Image {
	cfg, err := empty.Image.ConfigFile()
	Expect(err).ToNot(HaveOccurred())

	cfg.Author = "base"
	cfg.Config.Cmd = []string{"/bin/sh"}
	cfg.Config.Env = []string{"BASE=1"}
	cfg.Config.Labels = map[string]string{"base": "true"}
	cfg.Config.WorkingDir = "/base"

	img, err := mutate.ConfigFile(empty.Image, cfg)
	Expect(err).ToNot(HaveOccurred())

	return img
}

func newPlatformSourceImage() v1.Image {
	cfg, err := empty.Image.ConfigFile()
	Expect(err).ToNot(HaveOccurred())

	cfg.OS = "linux"
	cfg.Architecture = "arm64"

	img, err := mutate.ConfigFile(empty.Image, cfg)
	Expect(err).ToNot(HaveOccurred())

	return img
}

func newDockerSaveTar(img v1.Image) []byte {
	ref, err := name.ParseReference("example.com/test:latest")
	Expect(err).ToNot(HaveOccurred())

	var buf bytes.Buffer
	Expect(tarball.Write(ref, img, &buf)).To(Succeed())

	return buf.Bytes()
}
