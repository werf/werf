package container_registry_extensions

import (
	"archive/tar"
	"bytes"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The manifest-only image is pushed to registries as the scratch from stage base, and every
// descendant stage inherits its layer. A zero-byte layer body is not a parseable tar and breaks
// consumers like dive with EOF, so the layer must always be a valid, non-empty tar archive.
var _ = Describe("manifest-only image", func() {
	It("has a single valid empty tar layer with a matching diffID", func() {
		img := NewManifestOnlyImage(map[string]string{"werf": "test"})

		layers, err := img.Layers()
		Expect(err).NotTo(HaveOccurred())
		Expect(layers).To(HaveLen(1))

		rc, err := layers[0].Uncompressed()
		Expect(err).NotTo(HaveOccurred())
		data, err := io.ReadAll(rc)
		Expect(err).NotTo(HaveOccurred())
		Expect(rc.Close()).To(Succeed())

		Expect(len(data)).To(BeNumerically(">", 0), "a valid tar archive is never zero bytes")

		tr := tar.NewReader(bytes.NewReader(data))
		_, err = tr.Next()
		Expect(err).To(Equal(io.EOF), "expected a well-formed empty tar with no entries")

		expectedDiffID, _, err := v1.SHA256(bytes.NewReader(data))
		Expect(err).NotTo(HaveOccurred())
		diffID, err := layers[0].DiffID()
		Expect(err).NotTo(HaveOccurred())
		Expect(diffID).To(Equal(expectedDiffID))

		cfg, err := img.ConfigFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.RootFS.DiffIDs).To(Equal([]v1.Hash{expectedDiffID}))
		Expect(cfg.Config.Labels).To(Equal(map[string]string{"werf": "test"}))
	})
})
