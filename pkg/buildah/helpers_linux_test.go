//go:build linux
// +build linux

package buildah

import (
	"encoding/json"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"go.podman.io/buildah/docker"
	"go.podman.io/image/v5/manifest"
	"go.podman.io/storage"
)

func newStoreWithImage(config docker.V2Image) (storage.Store, string) {
	ginkgo.GinkgoHelper()

	dir := ginkgo.GinkgoT().TempDir()
	store, err := storage.GetStore(storage.StoreOptions{
		RunRoot: filepath.Join(dir, "run"), GraphRoot: filepath.Join(dir, "graph"), GraphDriverName: "vfs",
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(func() {
		_, err := store.Shutdown(false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	configBytes, err := json.Marshal(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	configDigest := digest.FromBytes(configBytes)
	manifestBytes, err := manifest.Schema2FromComponents(manifest.Schema2Descriptor{
		MediaType: manifest.DockerV2Schema2ConfigMediaType,
		Size:      int64(len(configBytes)),
		Digest:    configDigest,
	}, nil).Serialize()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	const ref = "registry.example.org/project/base:tag"
	created, err := store.CreateImage(configDigest.Encoded(), []string{ref}, "", "", &storage.ImageOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	computeDigest := func(data []byte) (digest.Digest, error) { return digest.FromBytes(data), nil }
	gomega.Expect(store.SetImageBigData(created.ID, configDigest.String(), configBytes, computeDigest)).To(gomega.Succeed())
	gomega.Expect(store.SetImageBigData(created.ID, storage.ImageDigestBigDataKey, manifestBytes, computeDigest)).To(gomega.Succeed())

	return store, ref
}

func dockerConfig(config docker.Config) docker.V2Image {
	return docker.V2Image{V1Image: docker.V1Image{Config: &config}}
}

func mustPushCompression(format string, level *int) pushCompression {
	ginkgo.GinkgoHelper()

	compressionOpts, err := newPushCompression(format, level)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return compressionOpts
}
