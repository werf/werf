package bundles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

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
				Raw: []*chart.File{
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

			Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())
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
			Raw: []*chart.File{
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

		Expect(from.CopyTo(ctx, to, copyToOptions{HelmCompatibleChart: true})).To(Succeed())

		remoteChart := bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()]
		Expect(remoteChart).NotTo(BeNil())

		VerifyChart(ctx, remoteChart, VerifyChartOptions{
			ExpectedName:    "testproject",
			ExpectedVersion: "1.2.3",
			ExpectedRepo:    "registry.example.com/group/testproject",
			ExpectedImages: map[string]string{
				"image-1": "registry.example.com/group/testproject:tag-1",
				"image-2": "registry.example.com/group/testproject:tag-2",
				"image-3": "registry.example.com/group/testproject:tag-3",
			},
		})

		{
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"]).To(Equal([]byte(`image-1-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"]).To(Equal([]byte(`image-2-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"]).To(Equal([]byte(`image-3-bytes`)))
		}
	})

	It("should copy archive to remote (without chart rename)", func() {
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
			Raw: []*chart.File{
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

		Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())

		remoteChart := bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()]
		Expect(remoteChart).NotTo(BeNil())

		VerifyChart(ctx, remoteChart, VerifyChartOptions{
			ExpectedName:    "test-bundle",
			ExpectedVersion: "1.2.3",
			ExpectedRepo:    "registry.example.com/group/testproject",
			ExpectedImages: map[string]string{
				"image-1": "registry.example.com/group/testproject:tag-1",
				"image-2": "registry.example.com/group/testproject:tag-2",
				"image-3": "registry.example.com/group/testproject:tag-3",
			},
		})

		{
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"]).To(Equal([]byte(`image-1-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"]).To(Equal([]byte(`image-2-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"]).To(Equal([]byte(`image-3-bytes`)))
		}
	})

	It("should copy archive to remote (without chart rename)", func() {
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
			Raw: []*chart.File{
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

		Expect(from.CopyTo(ctx, to, copyToOptions{RenameChart: "new-chart-name"})).To(Succeed())

		remoteChart := bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()]
		Expect(remoteChart).NotTo(BeNil())

		VerifyChart(ctx, remoteChart, VerifyChartOptions{
			ExpectedName:    "new-chart-name",
			ExpectedVersion: "1.2.3",
			ExpectedRepo:    "registry.example.com/group/testproject",
			ExpectedImages: map[string]string{
				"image-1": "registry.example.com/group/testproject:tag-1",
				"image-2": "registry.example.com/group/testproject:tag-2",
				"image-3": "registry.example.com/group/testproject:tag-3",
			},
		})

		{
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"]).To(Equal([]byte(`image-1-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"]).To(Equal([]byte(`image-2-bytes`)))
			Expect(registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"]).To(Equal([]byte(`image-3-bytes`)))
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
			Raw: []*chart.File{
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

		addr, err := ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()
		from := NewRemoteBundle(addr.RegistryAddress, bundlesRegistryClient, registryClient)

		bundlesRegistryClient.StubCharts[addr.RegistryAddress.FullName()] = ch
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"] = []byte(`image-2-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"] = []byte(`image-3-bytes`)

		toArchiveReaderStub := NewBundleArchiveStubReader(ch, nil)
		toArchiveWriterStub := NewBundleArchiveStubWriter()
		to := NewBundleArchive(toArchiveReaderStub, toArchiveWriterStub)

		Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())

		newCh := toArchiveWriterStub.StubChart
		Expect(newCh).NotTo(BeNil())

		origRepo := ch.Values["werf"].(map[string]interface{})["repo"].(string)

		VerifyChart(ctx, newCh, VerifyChartOptions{
			ExpectedName:    "testproject",
			ExpectedVersion: "1.2.3",
			ExpectedRepo:    origRepo,
			ExpectedImages: map[string]string{
				"image-1": "registry.example.com/group/testproject:tag-1",
				"image-2": "registry.example.com/group/testproject:tag-2",
				"image-3": "registry.example.com/group/testproject:tag-3",
			},
		})

		{
			Expect(toArchiveWriterStub.ImagesByTag["tag-1"]).To(Equal([]byte(`image-1-bytes`)))
			Expect(toArchiveWriterStub.ImagesByTag["tag-2"]).To(Equal([]byte(`image-2-bytes`)))
			Expect(toArchiveWriterStub.ImagesByTag["tag-3"]).To(Equal([]byte(`image-3-bytes`)))
		}
	})

	It("should copy remote to remote", func() {
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
			Raw: []*chart.File{
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

		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()

		fromAddr, err := ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
		bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"] = []byte(`image-2-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"] = []byte(`image-3-bytes`)

		toAddr, err := ParseAddr("registry2.example.com/group2/testproject2:4.5.6")
		Expect(err).NotTo(HaveOccurred())
		to := NewRemoteBundle(toAddr.RegistryAddress, bundlesRegistryClient, registryClient)

		Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())

		newCh := bundlesRegistryClient.StubCharts[toAddr.RegistryAddress.FullName()]
		Expect(newCh).NotTo(BeNil())

		VerifyChart(ctx, newCh, VerifyChartOptions{
			ExpectedName:    "testproject2",
			ExpectedVersion: "4.5.6",
			ExpectedRepo:    "registry2.example.com/group2/testproject2",
			ExpectedImages: map[string]string{
				"image-1": "registry2.example.com/group2/testproject2:tag-1",
				"image-2": "registry2.example.com/group2/testproject2:tag-2",
				"image-3": "registry2.example.com/group2/testproject2:tag-3",
			},
		})

		{
			Expect(registryClient.ImagesByReference["registry2.example.com/group2/testproject2:tag-1"]).To(Equal([]byte(`image-1-bytes`)))
			Expect(registryClient.ImagesByReference["registry2.example.com/group2/testproject2:tag-2"]).To(Equal([]byte(`image-2-bytes`)))
			Expect(registryClient.ImagesByReference["registry2.example.com/group2/testproject2:tag-3"]).To(Equal([]byte(`image-3-bytes`)))
		}
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

	ImagesByReference map[string][]byte
}

func NewDockerRegistryStub() *DockerRegistryStub {
	return &DockerRegistryStub{
		ImagesByReference: make(map[string][]byte),
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

	registry.ImagesByReference[reference] = buf.Bytes()

	return nil
}

func (registry *DockerRegistryStub) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error {
	data, hasImage := registry.ImagesByReference[reference]
	if !hasImage {
		return fmt.Errorf("image not found")
	}

	if _, err := io.Copy(archiveWriter, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("error copying image archive: %w", err)
	}
	return nil
}

func (registry *DockerRegistryStub) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) error {
	data, hasImage := registry.ImagesByReference[sourceReference]
	if !hasImage {
		return fmt.Errorf("source image not found")
	}

	registry.ImagesByReference[destinationReference] = data
	return nil
}

type VerifyChartOptions struct {
	ExpectedName    string
	ExpectedVersion string
	ExpectedRepo    string
	ExpectedImages  map[string]string
}

func VerifyChart(ctx context.Context, ch *chart.Chart, opts VerifyChartOptions) {
	Expect(ch.Metadata.Name).To(Equal(opts.ExpectedName))
	Expect(ch.Metadata.Version).To(Equal(opts.ExpectedVersion))

	werfVals := ch.Values["werf"].(map[string]interface{})

	repoVal := werfVals["repo"].(string)
	Expect(repoVal).To(Equal(opts.ExpectedRepo))

	imagesVals := werfVals["image"].(map[string]interface{})

	// No excess images in values, only expected
	for imageName := range imagesVals {
		Expect(opts.ExpectedImages[imageName]).NotTo(BeNil())
	}
	// All expected images are in values
	for imageName, imageRef := range opts.ExpectedImages {
		Expect(imagesVals[imageName]).NotTo(BeNil())
		Expect(imagesVals[imageName].(string)).To(Equal(imageRef))
	}

	chValues, err := json.Marshal(ch.Values)
	Expect(err).NotTo(HaveOccurred())

	var valuesFile *chart.File
	for _, f := range ch.Raw {
		if f.Name == "values.yaml" {
			valuesFile = f
			break
		}
	}
	Expect(valuesFile).NotTo(BeNil())

	var unmarshalledValues map[string]interface{}

	Expect(yaml.Unmarshal(valuesFile.Data, &unmarshalledValues)).To(Succeed())

	remarshalledValues, err := json.Marshal(unmarshalledValues)
	Expect(err).NotTo(HaveOccurred())

	fmt.Printf("Check chart loaded values:\n%s\n---\n", chValues)
	fmt.Printf("Check chart values.yaml:\n%s\n---\n", remarshalledValues)
	Expect(string(chValues)).To(Equal(string(remarshalledValues)))
}
