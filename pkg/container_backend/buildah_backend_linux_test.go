package container_backend

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/samber/lo"
	"go.podman.io/storage"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/buildah/thirdparty"
	"github.com/werf/werf/v3/test/pkg/buildahstub"
)

var _ = ginkgo.Describe("Buildah image metadata after retagging", func() {
	ginkgo.It("keeps the cached platform manifest digest when another platform takes the tag", func() {
		dir := ginkgo.GinkgoT().TempDir()
		store, err := storage.GetStore(storage.StoreOptions{
			RunRoot: filepath.Join(dir, "run"), GraphRoot: filepath.Join(dir, "graph"), GraphDriverName: "vfs",
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(func() {
			_, err := store.Shutdown(false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		const ref = "registry.example.org/project/base:tag"
		imageIDs := make(map[string]string)
		manifestDigests := make(map[string]string)
		for _, architecture := range []string{"amd64", "arm64"} {
			config, err := json.Marshal(v1.Image{
				Platform: v1.Platform{Architecture: architecture, OS: "linux"},
				RootFS:   v1.RootFS{Type: "layers", DiffIDs: []digest.Digest{}},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			configDigest := digest.FromBytes(config)
			manifest, err := json.Marshal(v1.Manifest{
				Versioned: specs.Versioned{SchemaVersion: 2},
				MediaType: v1.MediaTypeImageManifest,
				Config:    v1.Descriptor{MediaType: v1.MediaTypeImageConfig, Digest: configDigest, Size: int64(len(config))},
				Layers:    []v1.Descriptor{},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			created, err := store.CreateImage(configDigest.Encoded(), nil, "", "", &storage.ImageOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			computeDigest := func(data []byte) (digest.Digest, error) { return digest.FromBytes(data), nil }
			gomega.Expect(store.SetImageBigData(created.ID, configDigest.String(), config, computeDigest)).To(gomega.Succeed())
			gomega.Expect(store.SetImageBigData(created.ID, storage.ImageDigestBigDataKey, manifest, computeDigest)).To(gomega.Succeed())
			imageIDs[architecture] = created.ID
			manifestDigests[architecture] = digest.FromBytes(manifest).String()
		}
		gomega.Expect(store.AddNames(imageIDs["amd64"], []string{ref})).To(gomega.Succeed())
		native := &buildah.NativeBuildah{Store: store}
		fakeBuildah := &buildahstub.BuildahStub{ImagesFunc: native.Images}
		fakeBuildah.InspectFunc = func(_ context.Context, ref string) (*thirdparty.BuilderInfo, error) {
			stored, err := store.Image(strings.TrimPrefix(ref, "sha256:"))
			if err != nil {
				return nil, err
			}
			inspect := &thirdparty.BuilderInfo{FromImageID: stored.ID, FromImageDigest: stored.Digest.String()}
			inspect.Docker.Config = lo.ToPtr(lo.FromPtr(inspect.Docker.Config))
			return inspect, nil
		}
		backend := NewBuildahBackend(fakeBuildah, BuildahBackendOptions{})
		backend.storePulledImageID(ref, "linux/amd64", "sha256:"+imageIDs["amd64"])
		before, err := backend.GetImageInfo(context.Background(), ref, GetImageInfoOpts{TargetPlatform: "linux/amd64"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(before.RepoDigest).To(gomega.Equal("registry.example.org/project/base@" + manifestDigests["amd64"]))
		gomega.Expect(store.AddNames(imageIDs["arm64"], []string{ref})).To(gomega.Succeed())
		tagless, err := store.Image(imageIDs["amd64"])
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tagless.Names).To(gomega.BeEmpty())
		current, err := store.Image(ref)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(current.ID).To(gomega.Equal(imageIDs["arm64"]))
		after, err := backend.GetImageInfo(context.Background(), ref, GetImageInfoOpts{TargetPlatform: "linux/amd64"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(after.ID).To(gomega.Equal(before.ID))
		gomega.Expect(after.RepoDigest).To(gomega.Equal(before.RepoDigest))
		gomega.Expect(after.RepoDigest).NotTo(gomega.Equal("registry.example.org/project/base@" + manifestDigests["arm64"]))
		for _, length := range []int{8, 12, 64} {
			byID, err := backend.GetImageInfo(context.Background(), imageIDs["amd64"][:length], GetImageInfoOpts{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(byID.ID).To(gomega.Equal(before.ID))
			gomega.Expect(byID.RepoDigest).To(gomega.BeEmpty())
		}
	})
})
