package bundles

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/nelm/v2/pkg/chartyaml"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/pkg/logging"
	bundles_registry "github.com/werf/werf/v3/pkg/ref"
)

const werfDataChartYAML = `# keep this comment
apiVersion: v2
name: testproject
version: 1.2.3
annotations:
  helm.sh/images: |
    - name: app
      image: registry.example.com/group/testproject:tag-1@sha256:aaa
    - name: author
      image: docker.io/library/nginx:1.27
`

const werfDataExtraValues = `
werf:
  env: production
  commit:
    hash: abc
  image:
    app: registry.example.com/group/testproject:tag-1
  repo: registry.example.com/group/testproject
global:
  werf:
    env: production
    repo: registry.example.com/group/testproject
    images:
      app:
        digest: sha256:aaa
        ref: registry.example.com/group/testproject:tag-1@sha256:aaa
        ref_tag: registry.example.com/group/testproject:tag-1
`

var _ = Describe("Bundle copy of a chart with .werf-data", func() {
	var ctx context.Context

	BeforeEach(func(ctx0 context.Context) {
		ctx = logging.WithLogger(ctx0)
	})

	DescribeTable("should point the image values and the helm.sh/images entries of the werf images at the new repo",
		func(copyToNewRepo func(ch *chart.Chart, bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub)) {
			bundlesRegistryClient := NewBundlesRegistryClientStub()
			registryClient := NewDockerRegistryStub()

			copyToNewRepo(newWerfDataChart(werfDataChartYAML, werfDataExtraValues), bundlesRegistryClient, registryClient)

			newCh := bundlesRegistryClient.StubCharts["registry2.example.com/group2/testproject2:4.5.6"]
			Expect(newCh).NotTo(BeNil())
			Expect(registryClient.ImagesByReference["registry2.example.com/group2/testproject2:tag-1"]).To(Equal([]byte(`image-1-bytes`)))

			Expect(newCh.Values).To(Equal(map[string]interface{}{"replicas": float64(1)}))

			data, err := werfdata.Parse(newCh.WerfDataFiles)
			Expect(err).To(Succeed())
			Expect(data.Metadata.Images).To(Equal([]string{"app"}))
			Expect(data.Metadata.Release).To(Equal(&werfdata.ReleaseMetadata{Name: "testproject-production", Namespace: "testproject-production", Env: "production"}))
			Expect(data.ExtraValues).To(Equal(parseTestValues(`
werf:
  env: production
  commit:
    hash: abc
  image:
    app: registry2.example.com/group2/testproject2:tag-1
  repo: registry2.example.com/group2/testproject2
global:
  werf:
    env: production
    repo: registry2.example.com/group2/testproject2
    images:
      app:
        digest: sha256:aaa
        image: registry2.example.com/group2/testproject2
        name: testproject2
        name_ref: testproject2:tag-1@sha256:aaa
        name_tag: testproject2:tag-1
        namespace: group2
        ref: registry2.example.com/group2/testproject2:tag-1@sha256:aaa
        ref_tag: registry2.example.com/group2/testproject2:tag-1
        registry: registry2.example.com
        repository: group2/testproject2
        repository_ref: group2/testproject2:tag-1@sha256:aaa
        repository_tag: group2/testproject2:tag-1
        tag: tag-1
        tag_digest: tag-1@sha256:aaa
`)))

			Expect(string(chartYAMLOf(newCh))).To(Equal(`# keep this comment
apiVersion: v2
name: testproject2
version: 4.5.6
annotations:
  helm.sh/images: |
    - name: author
      image: docker.io/library/nginx:1.27
    - name: app
      image: registry2.example.com/group2/testproject2:tag-1@sha256:aaa
`))
			Expect(newCh.Metadata.Name).To(Equal("testproject2"))
			Expect(newCh.Metadata.Version).To(Equal("4.5.6"))
			Expect(newCh.Metadata.Annotations[chartyaml.ImagesAnnotation]).To(ContainSubstring("registry2.example.com/group2/testproject2:tag-1@sha256:aaa"))
		},
		Entry("remote to remote", func(ch *chart.Chart, bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub) {
			fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
			Expect(err).NotTo(HaveOccurred())
			from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
			bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = ch
			registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)

			toAddr, err := bundles_registry.ParseAddr("registry2.example.com/group2/testproject2:4.5.6")
			Expect(err).NotTo(HaveOccurred())
			Expect(from.CopyTo(context.Background(), NewRemoteBundle(toAddr.RegistryAddress, bundlesRegistryClient, registryClient), copyToOptions{})).To(Succeed())
		}),
		Entry("archive to remote", func(ch *chart.Chart, bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub) {
			from := NewBundleArchive(NewBundleArchiveStubReader(ch, map[string][]byte{"tag-1": []byte(`image-1-bytes`)}), NewBundleArchiveStubWriter())

			toAddr, err := bundles_registry.ParseAddr("registry2.example.com/group2/testproject2:4.5.6")
			Expect(err).NotTo(HaveOccurred())
			Expect(from.CopyTo(context.Background(), NewRemoteBundle(toAddr.RegistryAddress, bundlesRegistryClient, registryClient), copyToOptions{})).To(Succeed())
		}),
	)

	DescribeTable("should keep the original signed chart only when the copy changes nothing",
		func(extraValues string, expectedSigned bool) {
			bundlesRegistryClient := NewBundlesRegistryClientStub()
			registryClient := NewDockerRegistryStub()

			fromAddr, err := bundles_registry.ParseAddr("registry.example.com/group/testproject:1.2.3")
			Expect(err).NotTo(HaveOccurred())
			from := NewRemoteBundle(fromAddr.RegistryAddress, bundlesRegistryClient, registryClient)
			bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = newWerfDataChart(werfDataChartYAML, extraValues)
			bundlesRegistryClient.SignedCharts[fromAddr.RegistryAddress.FullName()] = true
			registryClient.ImagesByReference["registry.example.com/group/testproject:tag-1"] = []byte(`image-1-bytes`)

			toRef, err := bundles_registry.ParseReference("registry2.example.com/group2/testproject")
			Expect(err).NotTo(HaveOccurred())
			toRef.Tag = ""
			Expect(from.CopyTo(ctx, NewRemoteBundle(&bundles_registry.RegistryAddress{Reference: toRef}, bundlesRegistryClient, registryClient), copyToOptions{})).To(Succeed())

			pushedRef := bundlesRegistryClient.PushedRefs[0]
			Expect(bundlesRegistryClient.SignedCharts[pushedRef]).To(Equal(expectedSigned))
		},
		Entry("unchanged: no werf values", "", true),
		Entry("changed: werf values", werfDataExtraValues, false),
	)
})

func newWerfDataChart(chartYAML, extraValues string) *chart.Chart {
	data := &werfdata.Data{
		Metadata: werfdata.Metadata{
			Release: &werfdata.ReleaseMetadata{Name: "testproject-production", Namespace: "testproject-production", Env: "production"},
		},
	}
	if extraValues != "" {
		data.ExtraValues = parseTestValues(extraValues)
		data.Metadata.Images = []string{"app"}
	}

	werfDataFiles, err := data.ChartFiles()
	Expect(err).To(Succeed())

	metadata := &chart.Metadata{}
	Expect(yaml.Unmarshal([]byte(chartYAML), metadata)).To(Succeed())

	return &chart.Chart{
		Metadata:      metadata,
		Values:        map[string]interface{}{"replicas": float64(1)},
		WerfDataFiles: werfDataFiles,
		Raw: []*chartcommon.File{
			{Name: "Chart.yaml", Data: []byte(chartYAML)},
			{Name: "values.yaml", Data: []byte("replicas: 1\n")},
		},
	}
}

func chartYAMLOf(ch *chart.Chart) []byte {
	for _, f := range ch.Raw {
		if f.Name == "Chart.yaml" {
			return f.Data
		}
	}

	return nil
}

func parseTestValues(data string) map[string]interface{} {
	vals := map[string]interface{}{}
	Expect(yaml.Unmarshal([]byte(data), &vals)).To(Succeed())

	return vals
}
