package bundles

import (
	"bytes"
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"
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
