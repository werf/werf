package artifact

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("updateFallbackIndex", func() {
	makeDesc := func(name, artifactType string, annotations map[string]string) v1.Descriptor {
		if annotations == nil {
			annotations = map[string]string{}
		}
		return v1.Descriptor{
			MediaType:    types.OCIManifestSchema1,
			Digest:       v1.Hash{Algorithm: "sha256", Hex: "0000000000000000000000000000000000000000000000000000000000000000"},
			Size:         42,
			ArtifactType: artifactType,
			Annotations:  annotations,
		}
	}

	It("should add first entry to empty index", func() {
		desc := makeDesc("image-a", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-a"})
		idx := updateFallbackIndex(empty.Index, desc, "type/sbom", "image-a")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(1))
		Expect(im.Manifests[0].Annotations[image.WerfImageNameAnnotation]).To(Equal("image-a"))
	})

	It("should replace existing entry with same artifactType and imageName", func() {
		oldDesc := makeDesc("image-a", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-a"})
		idx := updateFallbackIndex(empty.Index, oldDesc, "type/sbom", "image-a")

		newDesc := makeDesc("image-a", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-a", "checksum": "v2"})
		idx = updateFallbackIndex(idx, newDesc, "type/sbom", "image-a")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(1))
		Expect(im.Manifests[0].Annotations["checksum"]).To(Equal("v2"))
	})

	It("should keep entries with different artifactType", func() {
		sbomDesc := makeDesc("image-a", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-a"})
		idx := updateFallbackIndex(empty.Index, sbomDesc, "type/sbom", "image-a")

		otherDesc := makeDesc("image-a", "type/signature", map[string]string{image.WerfImageNameAnnotation: "image-a"})
		idx = updateFallbackIndex(idx, otherDesc, "type/signature", "image-a")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(2))
	})

	It("should keep entries for different imageNames with same artifactType", func() {
		descA := makeDesc("image-a", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-a"})
		idx := updateFallbackIndex(empty.Index, descA, "type/sbom", "image-a")

		descB := makeDesc("image-b", "type/sbom", map[string]string{image.WerfImageNameAnnotation: "image-b"})
		idx = updateFallbackIndex(idx, descB, "type/sbom", "image-b")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(2))
	})
})

var _ = Describe("newStaticIndex / Digest", func() {
	It("should compute deterministic digest for same content", func() {
		desc := v1.Descriptor{
			MediaType: types.OCIManifestSchema1,
			Digest:    v1.Hash{Algorithm: "sha256", Hex: "abc123"},
			Size:      100,
		}

		idx1 := newStaticIndex([]v1.Descriptor{desc})
		idx2 := newStaticIndex([]v1.Descriptor{desc})

		d1, err := idx1.Digest()
		Expect(err).To(Succeed())
		d2, err := idx2.Digest()
		Expect(err).To(Succeed())

		Expect(d1).To(Equal(d2))
	})

	It("should produce different digests for different content", func() {
		desc1 := v1.Descriptor{MediaType: types.OCIManifestSchema1, Size: 100}
		desc2 := v1.Descriptor{MediaType: types.OCIManifestSchema1, Size: 200}

		d1, _ := newStaticIndex([]v1.Descriptor{desc1}).Digest()
		d2, _ := newStaticIndex([]v1.Descriptor{desc2}).Digest()

		Expect(d1).ToNot(Equal(d2))
	})
})

var _ = Describe("Annotations", func() {
	It("should define WerfPlatformAnnotation constant", func() {
		Expect(image.WerfPlatformAnnotation).To(Equal("io.werf.target-platform"))
	})

	It("should preserve platform annotation through updateFallbackIndex", func() {
		desc := v1.Descriptor{
			MediaType:    types.OCIManifestSchema1,
			Digest:       v1.Hash{Algorithm: "sha256", Hex: "0000000000000000000000000000000000000000000000000000000000000000"},
			Size:         42,
			ArtifactType: "type/sbom",
			Annotations: map[string]string{
				image.WerfImageNameAnnotation: "my-app",
				image.WerfPlatformAnnotation:  "linux/amd64",
			},
		}

		idx := updateFallbackIndex(empty.Index, desc, "type/sbom", "my-app")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(1))
		Expect(im.Manifests[0].Annotations[image.WerfPlatformAnnotation]).To(Equal("linux/amd64"))
	})

	It("should be independent of imageName replacement filter", func() {
		descA := v1.Descriptor{
			MediaType:    types.OCIManifestSchema1,
			Digest:       v1.Hash{Algorithm: "sha256", Hex: "0000000000000000000000000000000000000000000000000000000000000000"},
			Size:         42,
			ArtifactType: "type/sbom",
			Annotations: map[string]string{
				image.WerfImageNameAnnotation: "app-a",
				image.WerfPlatformAnnotation:  "linux/amd64",
			},
		}

		descB := v1.Descriptor{
			MediaType:    types.OCIManifestSchema1,
			Digest:       v1.Hash{Algorithm: "sha256", Hex: "0000000000000000000000000000000000000000000000000000000000000000"},
			Size:         42,
			ArtifactType: "type/sbom",
			Annotations: map[string]string{
				image.WerfImageNameAnnotation: "app-b",
				image.WerfPlatformAnnotation:  "linux/arm64",
			},
		}

		idx := updateFallbackIndex(empty.Index, descA, "type/sbom", "app-a")
		idx = updateFallbackIndex(idx, descB, "type/sbom", "app-b")

		im, err := idx.IndexManifest()
		Expect(err).To(Succeed())
		Expect(im.Manifests).To(HaveLen(2))
		Expect(im.Manifests[0].Annotations[image.WerfPlatformAnnotation]).To(Equal("linux/amd64"))
		Expect(im.Manifests[1].Annotations[image.WerfPlatformAnnotation]).To(Equal("linux/arm64"))
	})
})
