package docker_registry

import (
	"context"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("api.PushManifestList manifest list format", func(children []manifestListChild, expectedIndexType types.MediaType) {
	ctx := context.Background()
	fixture := newWritableBearerRegistryFixture()
	DeferCleanup(fixture.server.Close)
	registry := newAPI(apiOptions{InsecureRegistry: true})
	repo := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"

	manifests, platforms := pushChildImages(ctx, registry, repo, children)

	listRef := repo + ":multiarch"
	Expect(registry.PushManifestList(ctx, listRef, ManifestListOptions{
		Manifests: manifests,
		Platforms: platforms,
	})).To(Succeed())

	desc, _, err := registry.getImageDesc(ctx, listRef)
	Expect(err).NotTo(HaveOccurred())
	Expect(desc.MediaType).To(Equal(expectedIndexType))

	ii, err := desc.ImageIndex()
	Expect(err).NotTo(HaveOccurred())
	im, err := ii.IndexManifest()
	Expect(err).NotTo(HaveOccurred())
	Expect(im.MediaType).To(Equal(expectedIndexType))
	Expect(im.Manifests).To(HaveLen(len(children)))

	for i, child := range children {
		Expect(im.Manifests[i].MediaType).To(Equal(child.expectedTypeOut))
		Expect(im.Manifests[i].Platform.String()).To(Equal(child.platform))

		childImage, err := ii.Image(im.Manifests[i].Digest)
		Expect(err).NotTo(HaveOccurred())
		layers, err := childImage.Layers()
		Expect(err).NotTo(HaveOccurred())
		Expect(layers).To(HaveLen(1))
		Expect(layers[0].MediaType()).To(Equal(child.layerType))
	}
},
	Entry("keeps the docker manifest list for docker children", []manifestListChild{dockerChild}, types.DockerManifestList),
	Entry("publishes an oci index for oci children", []manifestListChild{ociZstdChild}, types.OCIImageIndex),
	Entry("publishes an oci index for mixed children", []manifestListChild{dockerChild, ociZstdChild}, types.OCIImageIndex),
)

var _ = DescribeTable("api.MutateAndPushImage manifest list format", func(children []manifestListChild, expectedIndexType types.MediaType) {
	ctx := context.Background()
	fixture := newWritableBearerRegistryFixture()
	DeferCleanup(fixture.server.Close)
	registry := newAPI(apiOptions{InsecureRegistry: true})
	repo := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"

	var ii v1.ImageIndex = empty.Index
	for _, child := range children {
		img := buildChildImage(child)
		desc, err := partial.Descriptor(img)
		Expect(err).NotTo(HaveOccurred())
		cf, err := img.ConfigFile()
		Expect(err).NotTo(HaveOccurred())
		desc.Platform = cf.Platform()
		ii = mutate.AppendManifests(ii, mutate.IndexAddendum{Add: img, Descriptor: *desc})
	}
	ii = mutate.IndexMediaType(ii, types.DockerManifestList)

	sourceRef := repo + ":source"
	sourceParsedRef, err := name.ParseReference(sourceRef, registry.parseReferenceOptionsForHost(sourceRef)...)
	Expect(err).NotTo(HaveOccurred())
	Expect(registry.writeToRemote(ctx, sourceParsedRef, ii)).To(Succeed())

	destRef := repo + ":dest"
	Expect(registry.MutateAndPushImage(ctx, sourceRef, destRef)).To(Succeed())

	desc, _, err := registry.getImageDesc(ctx, destRef)
	Expect(err).NotTo(HaveOccurred())
	Expect(desc.MediaType).To(Equal(expectedIndexType))

	publishedIndex, err := desc.ImageIndex()
	Expect(err).NotTo(HaveOccurred())
	im, err := publishedIndex.IndexManifest()
	Expect(err).NotTo(HaveOccurred())
	for i, child := range children {
		Expect(im.Manifests[i].MediaType).To(Equal(child.expectedTypeOut))
	}
},
	Entry("keeps the docker manifest list for docker children", []manifestListChild{dockerChild}, types.DockerManifestList),
	Entry("publishes an oci index for oci children", []manifestListChild{ociZstdChild}, types.OCIImageIndex),
	Entry("publishes an oci index for mixed children", []manifestListChild{dockerChild, ociZstdChild}, types.OCIImageIndex),
)
