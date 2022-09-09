package bundles

import (
	"bytes"
	"context"
	"fmt"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"

	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/docker_registry"
)

var _ = Describe("Bundle copy", func() {
	It("should copy archive to archive", func() {
		{
			ctx := context.Background()

			ch := &chart.Chart{
				Metadata: &chart.Metadata{
					APIVersion: "v2",
					Name:       "test-bundle",
					Version:    "0.1.0",
					Type:       "application",
				},
				Values: map[string]interface{}{
					"werf": map[string]interface{}{
						"image": map[string]interface{}{
							"image-1": "REPO:tag-1",
							"image-2": "REPO:tag-2",
							"image-3": "REPO:tag-3",
						},
						"repo": "REPO",
					},
				},
				Files: []*chart.File{
					{
						Name: "values.yaml",
						Data: []byte(`
werf:
  image:
    image-1: REPO:tag-1
    image-2: REPO:tag-2
    image-3: REPO:tag-3
  repo: REPO
  `),
					},
				},
			}

			images := map[string][]byte{
				"tag-1": []byte(`image-1-bytes`),
				"tag-2": []byte(`image-2-bytes`),
				"tag-3": []byte(`image-3-bytes`),
			}

			fromArchiveReaderStub := NewBundleArchiveStubReader(ch, images)
			fromArchiveWriterStub := NewBundleArchiveStubWriter()
			toArchiveReaderStub := NewBundleArchiveStubReader(nil, nil)
			toArchiveWriterStub := NewBundleArchiveStubWriter()

			from := NewBundleArchive(fromArchiveReaderStub, fromArchiveWriterStub)
			to := NewBundleArchive(toArchiveReaderStub, toArchiveWriterStub)

			Expect(from.CopyTo(ctx, to)).To(Succeed())
			Expect(toArchiveWriterStub.StubChart.Metadata).To(Equal(fromArchiveReaderStub.StubChart.Metadata))
			Expect(toArchiveWriterStub.StubChart.Values).To(Equal(fromArchiveReaderStub.StubChart.Values))

			for imgName, imgData := range toArchiveWriterStub.ImagesByTag {
				Expect(string(fromArchiveReaderStub.ImagesByTag[imgName])).To(Equal(string(imgData)))
			}
		}
	})

	It("should copy archive to remote", func() {
		ctx := context.Background()

		ch := &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v2",
				Name:       "test-bundle",
				Version:    "0.1.0",
				Type:       "application",
			},
			Values: map[string]interface{}{
				"werf": map[string]interface{}{
					"image": map[string]interface{}{
						"image-1": "REPO:tag-1",
						"image-2": "REPO:tag-2",
						"image-3": "REPO:tag-3",
					},
					"repo": "REPO",
				},
			},
			Files: []*chart.File{
				{
					Name: "values.yaml",
					Data: []byte(`
werf:
  image:
    image-1: REPO:tag-1
    image-2: REPO:tag-2
    image-3: REPO:tag-3
  repo: REPO
`),
				},
			},
		}

		images := map[string][]byte{
			"tag-1": []byte(`image-1-bytes`),
			"tag-2": []byte(`image-2-bytes`),
			"tag-3": []byte(`image-3-bytes`),
		}

		fromArchiveReaderStub := NewBundleArchiveStubReader(ch, images)
		fromArchiveWriterStub := NewBundleArchiveStubWriter()
		from := NewBundleArchive(fromArchiveReaderStub, fromArchiveWriterStub)

		addr, err := ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()
		to := NewRemoteBundle(addr.RegistryAddress, bundlesRegistryClient, registryClient)

		Expect(from.CopyTo(ctx, to)).To(Succeed())

		remoteChart := bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()]
		Expect(remoteChart).NotTo(BeNil())
		Expect(remoteChart.Metadata.Name).To(Equal("testproject"))
		Expect(remoteChart.Metadata.Version).To(Equal("1.2.3"))

		origImages := ch.Values["werf"].(map[string]interface{})["image"].(map[string]interface{})

		if werfVals, ok := remoteChart.Values["werf"].(map[string]interface{}); ok {
			if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
				for imageName, v := range imageVals {
					imageRef, ok := v.(string)
					Expect(ok).To(BeTrue())

					{
						// No unknown images in values
						_, ok := origImages[imageName]
						Expect(ok).To(BeTrue())
					}

					ref, err := bundles_registry.ParseReference(imageRef)
					Expect(err).NotTo(HaveOccurred())
					Expect(ref.Repo).To(Equal("registry.example.com/group/testproject"))
				}

				// No lost images
				for imageName := range origImages {
					_, ok := imageVals[imageName]
					Expect(ok).To(BeTrue())
				}
			}

			Expect(werfVals["repo"]).To(Equal("registry.example.com/group/testproject"))
		}
	})

	It("should copy remote to archive", func() {
		ctx := context.Background()

		ch := &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v2",
				Name:       "testproject",
				Version:    "1.2.3",
				Type:       "application",
			},
			Values: map[string]interface{}{
				"werf": map[string]interface{}{
					"image": map[string]interface{}{
						"image-1": "registry.example.com/group/testproject:tag-1",
						"image-2": "registry.example.com/group/testproject:tag-2",
						"image-3": "registry.example.com/group/testproject:tag-3",
					},
					"repo": "registry.example.com/group/testproject",
				},
			},
			Files: []*chart.File{
				{
					Name: "values.yaml",
					Data: []byte(`
werf:
  image:
    image-1: registry.example.com/group/testproject:tag-1
    image-2: registry.example.com/group/testproject:tag-2
    image-3: registry.example.com/group/testproject:tag-3
  repo: registry.example.com/group/testproject
`),
				},
			},
		}

		images := map[string][]byte{
			"tag-1": []byte(`image-1-bytes`),
			"tag-2": []byte(`image-2-bytes`),
			"tag-3": []byte(`image-3-bytes`),
		}

		addr, err := ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()
		from := NewRemoteBundle(addr.RegistryAddress, bundlesRegistryClient, registryClient)

		bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()] = ch
		registryClient.RemoteImages["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)
		registryClient.RemoteImages["registry.example.com/group/testproject:tag-2"] = []byte(`image-2-bytes`)
		registryClient.RemoteImages["registry.example.com/group/testproject:tag-3"] = []byte(`image-3-bytes`)

		toArchiveReaderStub := NewBundleArchiveStubReader(ch, images)
		toArchiveWriterStub := NewBundleArchiveStubWriter()
		to := NewBundleArchive(toArchiveReaderStub, toArchiveWriterStub)

		Expect(from.CopyTo(ctx, to)).To(Succeed())
		Expect(toArchiveWriterStub.StubChart.Metadata.Name).To(Equal("testproject"))
		Expect(toArchiveWriterStub.StubChart.Metadata.Version).To(Equal("1.2.3"))

		origImages := ch.Values["werf"].(map[string]interface{})["image"].(map[string]interface{})
		origRepo := ch.Values["werf"].(map[string]interface{})["repo"].(string)
		newImages := toArchiveWriterStub.StubChart.Values["werf"].(map[string]interface{})["image"].(map[string]interface{})
		newRepo := toArchiveWriterStub.StubChart.Values["werf"].(map[string]interface{})["repo"].(string)

		for imageName, v := range origImages {
			Expect(newImages[imageName]).To(Equal(v))
		}
		for imageName, v := range newImages {
			Expect(origImages[imageName]).To(Equal(v))
		}
		Expect(newRepo).To(Equal(origRepo))
	})
})

type BundleArchiveStubReader struct {
	StubChart   *chart.Chart
	ImagesByTag map[string][]byte
}

func NewBundleArchiveStubReader(stubChart *chart.Chart, imagesByTag map[string][]byte) *BundleArchiveStubReader {
	return &BundleArchiveStubReader{StubChart: stubChart, ImagesByTag: imagesByTag}
}

func (reader *BundleArchiveStubReader) String() string {
	return "bundle-archive-stub-reader"
}

func (reader *BundleArchiveStubReader) ReadChartArchive() ([]byte, error) {
	return ChartToBytes(reader.StubChart)
}

func (reader *BundleArchiveStubReader) ReadImageArchive(imageTag string) (*ImageArchiveReadCloser, error) {
	data, hasTag := reader.ImagesByTag[imageTag]
	if !hasTag {
		return nil, fmt.Errorf("no image found by tag %q", imageTag)
	}

	return NewImageArchiveReadCloser(bytes.NewReader(data), func() error { return nil }), nil
}

type BundleArchiveStubWriter struct {
	StubChart   *chart.Chart
	ImagesByTag map[string][]byte
}

func NewBundleArchiveStubWriter() *BundleArchiveStubWriter {
	return &BundleArchiveStubWriter{ImagesByTag: make(map[string][]byte)}
}

func (writer *BundleArchiveStubWriter) Open() error { return nil }

func (writer *BundleArchiveStubWriter) WriteChartArchive(data []byte) error {
	ch, err := BytesToChart(data)
	if err != nil {
		return err
	}

	writer.StubChart = ch

	return nil
}

func (writer *BundleArchiveStubWriter) WriteImageArchive(imageTag string, data []byte) error {
	writer.ImagesByTag[imageTag] = data
	return nil
}

func (writer *BundleArchiveStubWriter) Save() error { return nil }

type BundlesRegistryClientStub struct {
	StubCharts map[string]*chart.Chart
}

func NewBundlesRegistryClientStub() *BundlesRegistryClientStub {
	return &BundlesRegistryClientStub{
		StubCharts: make(map[string]*chart.Chart),
	}
}

func (client *BundlesRegistryClientStub) PullChartToCache(ref *bundles_registry.Reference) error {
	return nil
}

func (client *BundlesRegistryClientStub) LoadChart(ref *bundles_registry.Reference) (*chart.Chart, error) {
	if ch, hasChart := client.StubCharts[ref.FullName()]; hasChart {
		return ch, nil
	}
	return nil, fmt.Errorf("no chart found by address %s", ref.FullName())
}

func (client *BundlesRegistryClientStub) SaveChart(ch *chart.Chart, ref *bundles_registry.Reference) error {
	client.StubCharts[ref.FullName()] = ch
	return nil
}

func (client *BundlesRegistryClientStub) PushChart(ref *bundles_registry.Reference) error {
	return nil
}

type DockerRegistryStub struct {
	docker_registry.Interface

	ImagesArchives map[string][]byte
	RemoteImages   map[string][]byte
}

func NewDockerRegistryStub() *DockerRegistryStub {
	return &DockerRegistryStub{
		ImagesArchives: make(map[string][]byte),
		RemoteImages:   make(map[string][]byte),
	}
}

func (registry *DockerRegistryStub) PushImageArchive(ctx context.Context, archiveOpener docker_registry.ArchiveOpener, reference string) error {
	readCloser, err := archiveOpener.Open()
	if err != nil {
		return fmt.Errorf("unable to open archive: %w", err)
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, readCloser); err != nil {
		return fmt.Errorf("error reading image archive: %w", err)
	}

	if err := readCloser.Close(); err != nil {
		return fmt.Errorf("unable to close image archive reader: %w", err)
	}

	registry.ImagesArchives[reference] = buf.Bytes()

	return nil
}

func (registry *DockerRegistryStub) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error {
	data, hasImage := registry.RemoteImages[reference]
	if !hasImage {
		return fmt.Errorf("image not found")
	}

	if _, err := io.Copy(archiveWriter, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("error copying image archive: %w", err)
	}
	return nil
}

func (registry *DockerRegistryStub) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) error {
	data, hasImage := registry.RemoteImages[sourceReference]
	if !hasImage {
		return fmt.Errorf("source image not found")
	}

	registry.RemoteImages[destinationReference] = data
	return nil
}
