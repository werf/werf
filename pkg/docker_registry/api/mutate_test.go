package api

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docker Registry Api Suite")
}

var _ = DescribeTable("mutateImage keeps the media types of the source image",
	func(manifestMediaType, configMediaType types.MediaType) {
		source, err := random.Image(64, 1)
		Expect(err).NotTo(HaveOccurred())
		source = mutate.MediaType(source, manifestMediaType)
		source = mutate.ConfigMediaType(source, configMediaType)

		dest, err := name.NewTag("registry.example.com/repo:tag")
		Expect(err).NotTo(HaveOccurred())

		mutated, _, err := mutateImage(context.Background(), source, dest, false,
			WithLayersMutation(func(_ context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
				addenda := make([]mutate.Addendum, 0, len(layers))
				for _, layer := range layers {
					addenda = append(addenda, mutate.Addendum{Layer: layer})
				}
				return addenda, nil
			}),
		)
		Expect(err).NotTo(HaveOccurred())

		manifest, err := mutated.Manifest()
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest.MediaType).To(Equal(manifestMediaType))
		Expect(manifest.Config.MediaType).To(Equal(configMediaType))
	},
	Entry("docker", types.DockerManifestSchema2, types.DockerConfigJSON),
	Entry("oci", types.OCIManifestSchema1, types.OCIConfigJSON),
)

var _ = It("mutateIndex publishes a docker manifest list", func() {
	index := mutate.AppendManifests(empty.Index)
	dest, err := name.NewTag("registry.example.com/repo:tag")
	Expect(err).NotTo(HaveOccurred())

	mutated, _, err := mutateIndex(context.Background(), index, dest, false)
	Expect(err).NotTo(HaveOccurred())
	Expect(mutated.MediaType()).To(Equal(types.DockerManifestList))
})
