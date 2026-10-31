package bundles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/logging"
	bundles_registry "github.com/werf/werf/v3/pkg/ref"
)

var _ = Describe("Bundle copy", func() {
	var ctx context.Context

	BeforeEach(func(ctx0 context.Context) {
		ctx = logging.WithLogger(ctx0)
	})

	It("should copy archive to archive", func() {
		{
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
							"image-1": "repo:tag-1",
							"image-2": "repo:tag-2",
							"image-3": "repo:tag-3",
						},
						"repo": "repo",
					},
				},
				Raw: []*chartcommon.File{
					{
						Name: "values.yaml",
						Data: []byte(`
werf:
  image:
    image-1: repo:tag-1
    image-2: repo:tag-2
    image-3: repo:tag-3
  repo: repo
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
						"image-1": "repo:tag-1",
						"image-2": "repo:tag-2",
						"image-3": "repo:tag-3",
					},
					"repo": "repo",
				},
			},
			Raw: []*chartcommon.File{
				{
					Name: "values.yaml",
					Data: []byte(`
werf:
  image:
    image-1: repo:tag-1
    image-2: repo:tag-2
    image-3: repo:tag-3
  repo: repo
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

		addr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()
		to := NewRemoteBundle(addr.RegistryAddress, bundlesRegistryClient, registryClient)

		Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())

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

	It("should copy remote to archive", func() {
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
			Raw: []*chartcommon.File{
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

		addr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
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
			Raw: []*chartcommon.File{
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

		fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
		bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-2"] = []byte(`image-2-bytes`)
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-3"] = []byte(`image-3-bytes`)

		toAddr, err := bundles_registry.ParseAddr("registry2.example.com/group2/testproject2:4.5.6")
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

	It("should copy remote to remote updating global werf values", func() {
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
						"myapp": "registry.example.com/group/testproject:tag-1",
					},
					"repo": "registry.example.com/group/testproject",
				},
				"global": map[string]interface{}{
					"werf": map[string]interface{}{
						"repo": "registry.example.com/group/testproject",
						"images": map[string]interface{}{
							"myapp": map[string]interface{}{
								"registry":       "registry.example.com",
								"namespace":      "group",
								"name":           "testproject",
								"tag":            "tag-1",
								"digest":         "sha256:abcdef1234567890",
								"tag_digest":     "tag-1@sha256:abcdef1234567890",
								"image":          "registry.example.com/group/testproject",
								"repository":     "group/testproject",
								"ref":            "registry.example.com/group/testproject:tag-1@sha256:abcdef1234567890",
								"ref_tag":        "registry.example.com/group/testproject:tag-1",
								"repository_ref": "group/testproject:tag-1@sha256:abcdef1234567890",
								"repository_tag": "group/testproject:tag-1",
								"name_ref":       "testproject:tag-1@sha256:abcdef1234567890",
								"name_tag":       "testproject:tag-1",
							},
						},
					},
				},
			},
			Raw: []*chartcommon.File{
				{
					Name: "values.yaml",
					Data: []byte(`
werf:
  image:
    myapp: registry.example.com/group/testproject:tag-1
  repo: registry.example.com/group/testproject
global:
  werf:
    repo: registry.example.com/group/testproject
    images:
      myapp:
        registry: registry.example.com
        namespace: group
        name: testproject
        tag: tag-1
        digest: "sha256:abcdef1234567890"
        tag_digest: "tag-1@sha256:abcdef1234567890"
        image: registry.example.com/group/testproject
        repository: group/testproject
        ref: "registry.example.com/group/testproject:tag-1@sha256:abcdef1234567890"
        ref_tag: "registry.example.com/group/testproject:tag-1"
        repository_ref: "group/testproject:tag-1@sha256:abcdef1234567890"
        repository_tag: "group/testproject:tag-1"
        name_ref: "testproject:tag-1@sha256:abcdef1234567890"
        name_tag: "testproject:tag-1"
`),
				},
			},
		}

		bundlesRegistryClient := NewBundlesRegistryClientStub()
		registryClient := NewDockerRegistryStub()

		fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
		Expect(err).NotTo(HaveOccurred())
		from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
		bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch
		registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)

		toAddr, err := bundles_registry.ParseAddr("registry2.example.com/org/newproject:4.5.6")
		Expect(err).NotTo(HaveOccurred())
		to := NewRemoteBundle(toAddr.RegistryAddress, bundlesRegistryClient, registryClient)

		Expect(from.CopyTo(ctx, to, copyToOptions{})).To(Succeed())

		newCh := bundlesRegistryClient.StubCharts[toAddr.RegistryAddress.FullName()]
		Expect(newCh).NotTo(BeNil())

		globalVals := newCh.Values["global"].(map[string]interface{})
		globalWerfVals := globalVals["werf"].(map[string]interface{})
		Expect(globalWerfVals["repo"]).To(Equal("registry2.example.com/org/newproject"))

		imagesVals := globalWerfVals["images"].(map[string]interface{})
		myappVals := imagesVals["myapp"].(map[string]interface{})

		Expect(myappVals["registry"]).To(Equal("registry2.example.com"))
		Expect(myappVals["namespace"]).To(Equal("org"))
		Expect(myappVals["name"]).To(Equal("newproject"))
		Expect(myappVals["tag"]).To(Equal("tag-1"))
		Expect(myappVals["digest"]).To(Equal("sha256:abcdef1234567890"))
		Expect(myappVals["tag_digest"]).To(Equal("tag-1@sha256:abcdef1234567890"))
		Expect(myappVals["image"]).To(Equal("registry2.example.com/org/newproject"))
		Expect(myappVals["repository"]).To(Equal("org/newproject"))
		Expect(myappVals["ref"]).To(Equal("registry2.example.com/org/newproject:tag-1@sha256:abcdef1234567890"))
		Expect(myappVals["ref_tag"]).To(Equal("registry2.example.com/org/newproject:tag-1"))
		Expect(myappVals["repository_ref"]).To(Equal("org/newproject:tag-1@sha256:abcdef1234567890"))
		Expect(myappVals["repository_tag"]).To(Equal("org/newproject:tag-1"))
		Expect(myappVals["name_ref"]).To(Equal("newproject:tag-1@sha256:abcdef1234567890"))
		Expect(myappVals["name_tag"]).To(Equal("newproject:tag-1"))
	})

	It("should copy archive to remote without a tag keeping the chart version", func() {
		ch := &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v2",
				Name:       "testproject",
				Version:    "0.1.0",
				Type:       "application",
			},
			Values: map[string]interface{}{
				"werf": map[string]interface{}{
					"image": map[string]interface{}{},
					"repo":  "repo",
				},
			},
			Raw: []*chartcommon.File{
				{
					Name: "values.yaml",
					Data: []byte("werf:\n  image: {}\n  repo: repo\n"),
				},
			},
		}

		from := NewBundleArchive(NewBundleArchiveStubReader(ch, map[string][]byte{}), NewBundleArchiveStubWriter())

		bundlesRegistryClient := NewBundlesRegistryClientStub()
		toAddr := &bundles_registry.RegistryAddress{Reference: &bundles_registry.Reference{Repo: "registry.example.com/group/testproject"}}
		Expect(from.CopyTo(ctx, NewRemoteBundle(toAddr, bundlesRegistryClient, NewDockerRegistryStub()), copyToOptions{})).To(Succeed())

		Expect(bundlesRegistryClient.PushedRefs).To(Equal([]string{"registry.example.com/group/testproject:0.1.0"}))
		Expect(bundlesRegistryClient.StubCharts["registry.example.com/group/testproject:0.1.0"].Metadata.Version).To(Equal("0.1.0"))
	})

	DescribeTable("should copy a signed remote bundle to remote keeping the original chart only when the copy changes nothing",
		func(to string, values map[string]interface{}, expectedSigned bool) {
			ch := &chart.Chart{
				Metadata: &chart.Metadata{
					APIVersion: "v2",
					Name:       "signedchart",
					Version:    "0.1.0",
				},
				Values: values,
				Raw:    []*chartcommon.File{{Name: "values.yaml", Data: []byte("{}\n")}},
			}

			bundlesRegistryClient := NewBundlesRegistryClientStub()
			registryClient := NewDockerRegistryStub()

			fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/signedchart:0.1.0")
			Expect(err).NotTo(HaveOccurred())
			from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
			bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch
			bundlesRegistryClient.SignedCharts[fromAddr.RegistryAddress.FullName()] = true

			toRef, err := bundles_registry.ParseReference(to)
			Expect(err).NotTo(HaveOccurred())
			if !strings.Contains(to, ":") {
				toRef.Tag = ""
			}
			Expect(from.CopyTo(ctx, NewRemoteBundle(&bundles_registry.RegistryAddress{Reference: toRef}, bundlesRegistryClient, registryClient), copyToOptions{})).To(Succeed())

			Expect(bundlesRegistryClient.PushedRefs).NotTo(BeEmpty())
			pushedRef := bundlesRegistryClient.PushedRefs[0]
			Expect(bundlesRegistryClient.SignedCharts[pushedRef]).To(Equal(expectedSigned))
			if expectedSigned {
				Expect(bundlesRegistryClient.TaggedCachedRefs).To(HaveKeyWithValue(pushedRef, fromAddr.RegistryAddress.FullName()))
			} else {
				Expect(bundlesRegistryClient.TaggedCachedRefs).To(BeEmpty())
			}
		},
		Entry("unchanged: another repository with the same name, no tag",
			"registry2.example.com/group2/signedchart", map[string]interface{}{"replicas": 1}, true),
		Entry("unchanged: a tag equal to the version",
			"registry2.example.com/group2/signedchart:0.1.0", map[string]interface{}{"replicas": 1}, true),
		Entry("changed version",
			"registry2.example.com/group2/signedchart:0.2.0", map[string]interface{}{"replicas": 1}, false),
		Entry("changed name",
			"registry2.example.com/group2/otherchart", map[string]interface{}{"replicas": 1}, false),
		Entry("changed werf values",
			"registry2.example.com/group2/signedchart", map[string]interface{}{"werf": map[string]interface{}{"repo": "registry.example.com/group/signedchart"}}, false),
	)

	DescribeTable("should copy remote to remote under the version and the given tag",
		func(fromTag, fromVersion, to, expectedVersion string, expectedPushedRefs []string) {
			ch := &chart.Chart{
				Metadata: &chart.Metadata{
					APIVersion: "v2",
					Name:       "testproject",
					Version:    fromVersion,
					Type:       "application",
				},
				Values: map[string]interface{}{
					"werf": map[string]interface{}{
						"image": map[string]interface{}{},
						"repo":  "registry.example.com/group/testproject",
					},
				},
				Raw: []*chartcommon.File{
					{
						Name: "values.yaml",
						Data: []byte("werf:\n  image: {}\n  repo: registry.example.com/group/testproject\n"),
					},
				},
			}

			bundlesRegistryClient := NewBundlesRegistryClientStub()
			registryClient := NewDockerRegistryStub()

			fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:" + fromTag)
			Expect(err).NotTo(HaveOccurred())
			from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
			bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch

			toAddr := &bundles_registry.RegistryAddress{Reference: &bundles_registry.Reference{Repo: "registry2.example.com/group2/testproject2", Tag: to}}
			Expect(from.CopyTo(ctx, NewRemoteBundle(toAddr, bundlesRegistryClient, registryClient), copyToOptions{})).To(Succeed())

			savedChart := bundlesRegistryClient.StubCharts[bundlesRegistryClient.PushedRefs[0]]
			Expect(savedChart).NotTo(BeNil())
			Expect(savedChart.Metadata.Version).To(MatchRegexp(expectedVersion))
			Expect(bundlesRegistryClient.PushedRefs).To(Equal(lo.Map(expectedPushedRefs, func(pushedRef string, _ int) string {
				return strings.ReplaceAll(pushedRef, "{version}", savedChart.Metadata.Version)
			})))
		},
		Entry("semver tag", "1.2.3", "1.2.3", "4.5.6", `^4\.5\.6$`,
			[]string{"registry2.example.com/group2/testproject2:4.5.6"}),
		Entry("semver tag with the v prefix", "1.2.3", "1.2.3", "v4.5.6", `^4\.5\.6$`,
			[]string{"registry2.example.com/group2/testproject2:4.5.6", "registry2.example.com/group2/testproject2:v4.5.6"}),
		Entry("non-semver tag", "1.2.3", "1.2.3", "latest", `^0\.0\.0-\d+-latest$`,
			[]string{"registry2.example.com/group2/testproject2:{version}", "registry2.example.com/group2/testproject2:latest"}),
		Entry("no tag keeps the source version and tag", "v1.2.3", "1.2.3", "", `^1\.2\.3$`,
			[]string{"registry2.example.com/group2/testproject2:1.2.3", "registry2.example.com/group2/testproject2:v1.2.3"}),
		Entry("no tag keeps the source version with build metadata", "1.2.3", "1.2.3+build.5", "", `^1\.2\.3\+build\.5$`,
			[]string{"registry2.example.com/group2/testproject2:1.2.3_build.5", "registry2.example.com/group2/testproject2:1.2.3"}),
	)
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

func (writer *BundleArchiveStubWriter) WriteChartArchive(data []byte, opts nelmcommon.HelmOptions) error {
	ch, err := BytesToChart(context.Background(), data, opts)
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
	StubCharts       map[string]*chart.Chart
	SignedCharts     map[string]bool
	PushedRefs       []string
	TaggedCachedRefs map[string]string
}

func NewBundlesRegistryClientStub() *BundlesRegistryClientStub {
	return &BundlesRegistryClientStub{
		StubCharts:       make(map[string]*chart.Chart),
		SignedCharts:     make(map[string]bool),
		TaggedCachedRefs: make(map[string]string),
	}
}

func (client *BundlesRegistryClientStub) PullChartToCache(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error {
	return nil
}

func (client *BundlesRegistryClientStub) LoadChart(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) (*chart.Chart, error) {
	if ch, hasChart := client.StubCharts[ref.FullName()]; hasChart {
		return ch, nil
	}
	return nil, fmt.Errorf("no chart found by address %s", ref.FullName())
}

func (client *BundlesRegistryClientStub) SaveChartArchive(ctx context.Context, chartArchive []byte, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error {
	ch, err := BytesToChart(ctx, chartArchive, opts)
	if err != nil {
		return err
	}

	client.StubCharts[ref.FullName()] = ch
	return nil
}

func (client *BundlesRegistryClientStub) PushChart(ctx context.Context, ref *bundles_registry.Reference, tags []string, opts nelmcommon.HelmOptions) error {
	if _, hasChart := client.StubCharts[ref.FullName()]; !hasChart {
		return fmt.Errorf("no chart saved by address %s", ref.FullName())
	}

	for _, tag := range tags {
		client.PushedRefs = append(client.PushedRefs, fmt.Sprintf("%s:%s", ref.Repo, tag))
	}

	return nil
}

func (client *BundlesRegistryClientStub) TagCachedChart(ctx context.Context, from, to *bundles_registry.Reference, opts nelmcommon.HelmOptions) error {
	ch, hasChart := client.StubCharts[from.FullName()]
	if !hasChart {
		return fmt.Errorf("no chart found by address %s", from.FullName())
	}

	client.StubCharts[to.FullName()] = ch
	client.SignedCharts[to.FullName()] = client.SignedCharts[from.FullName()]
	client.TaggedCachedRefs[to.FullName()] = from.FullName()

	return nil
}

func (client *BundlesRegistryClientStub) IsChartSigned(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) (bool, error) {
	return client.SignedCharts[ref.FullName()], nil
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

func (registry *DockerRegistryStub) CopyImage(_ context.Context, sourceReference, destinationReference string, _ docker_registry.CopyImageOptions) error {
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

	var valuesFile *chartcommon.File
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
