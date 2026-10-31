package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/chartyaml"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("Bundle copy", func() {
	var (
		ctx                context.Context
		fromRepo, toRepo   string
		imageDigest        string
		copyOpts           CopyOptions
		testRegistryClient *registry.Client
	)

	BeforeEach(func() {
		ctx = logging.WithLogger(context.Background())
		registryHost := startTestRegistry()
		fromRepo = registryHost + "/group/testproject"
		toRepo = registryHost + "/org/newproject"
		imageDigest = pushTestImage(fromRepo + ":tag-1")
		testRegistryClient = newTestRegistryClient(ctx)
		copyOpts = CopyOptions{
			ChartRepoConnectionOptions: nelmcommon.ChartRepoConnectionOptions{ChartRepoInsecure: true},
			RegistryCredentialsPath:    testCredentialsPath(),
		}
	})

	DescribeTable("should copy a bundle published by werf 3.12 or older with its images and point its image values at the new repo",
		func(copyBundle func(from, to string)) {
			pushOldBundle(testRegistryClient, fromRepo, "1.2.3", oldBundleValues(fromRepo, imageDigest))

			copyBundle(fromRepo+":1.2.3", toRepo+":4.5.6")

			Expect(testImageDigest(toRepo + ":tag-1")).To(Equal(imageDigest))

			ch := pullTestChart(ctx, testRegistryClient, toRepo, "4.5.6")
			Expect(ch.Metadata.Name).To(Equal("newproject"))
			Expect(ch.WerfDataFiles).To(BeEmpty())
			Expect(ch.Values).To(Equal(oldBundleValues(toRepo, imageDigest)))
		},
		Entry("registry to registry", func(from, to string) {
			Expect(Copy(ctx, from, to, copyOpts)).To(Succeed())
		}),
		Entry("registry to archive to registry", func(from, to string) {
			archivePath := filepath.Join(GinkgoT().TempDir(), "bundle.tar.gz")
			Expect(Copy(ctx, from, "archive:"+archivePath, copyOpts)).To(Succeed())
			Expect(Copy(ctx, "archive:"+archivePath, to, copyOpts)).To(Succeed())
		}),
	)

	It("should point the image values of a bundle imported from a werf 3.12 archive at the imported images", func() {
		img, err := random.Image(64, 1)
		Expect(err).To(Succeed())

		// The archive keeps the image layers but not its manifest, which the import rebuilds.
		staleDigest := "sha256:" + strings.Repeat("a", 64)
		archivePath := filepath.Join(GinkgoT().TempDir(), "bundle.tar.gz")
		writeWerf3Archive(archivePath, oldBundleArchive(fromRepo, "1.2.3", oldBundleValues(fromRepo, staleDigest)), map[string]v1.Image{"tag-1": img})

		Expect(Copy(ctx, "archive:"+archivePath, toRepo, copyOpts)).To(Succeed())

		ch := pullTestChart(ctx, testRegistryClient, toRepo, "1.2.3")
		images := ch.Values["global"].(map[string]interface{})["werf"].(map[string]interface{})["images"].(map[string]interface{})
		Expect(images).NotTo(BeEmpty())
		for name, imageVals := range images {
			imageRef := imageVals.(map[string]interface{})["ref"].(string)
			Expect(imageRef).To(HavePrefix(toRepo+":tag-1@"), name)
			testImageDigest(imageRef)
		}
	})

	It("should copy between plain HTTP registries on loopback hosts and archives without insecure", func() {
		copyOpts.ChartRepoConnectionOptions = nelmcommon.ChartRepoConnectionOptions{}
		pushOldBundle(testRegistryClient, fromRepo, "1.2.3", oldBundleValues(fromRepo, imageDigest))

		archivePath := filepath.Join(GinkgoT().TempDir(), "bundle.tar")
		Expect(Copy(ctx, fromRepo+":1.2.3", "archive:"+archivePath, copyOpts)).To(Succeed())
		Expect(Copy(ctx, "archive:"+archivePath, toRepo, copyOpts)).To(Succeed())

		Expect(testImageDigest(toRepo + ":tag-1")).To(Equal(imageDigest))
	})

	DescribeTable("should copy a chart with .werf-data with its images and point its image values at the new repo",
		func(copyBundle func(from, to string)) {
			Expect(Publish(ctx, writeTestChart(), fromRepo, "1.2.3", testRegistryClient, PublishOptions{
				ChartPackOptions: werfDataChartPackOptions(ctx, fromRepo, imageDigest),
			})).To(Succeed())

			copyBundle(fromRepo+":1.2.3", toRepo+":4.5.6")

			Expect(testImageDigest(toRepo + ":tag-1")).To(Equal(imageDigest))

			ch := pullTestChart(ctx, testRegistryClient, toRepo, "4.5.6")
			Expect(ch.Metadata.Name).To(Equal("newproject"))
			Expect(ch.Values).To(BeEmpty())
			Expect(ch.Metadata.Annotations[chartyaml.ImagesAnnotation]).To(ContainSubstring(toRepo + ":tag-1@" + imageDigest))

			data, err := werfdata.Parse(ch.WerfDataFiles)
			Expect(err).To(Succeed())
			Expect(data.Metadata.Images).To(Equal([]string{"app"}))
			Expect(data.ExtraValues).To(Equal(werfDataExtraValues(toRepo, imageDigest)))
		},
		Entry("registry to registry", func(from, to string) {
			Expect(Copy(ctx, from, to, copyOpts)).To(Succeed())
		}),
		Entry("registry to archive to registry", func(from, to string) {
			archivePath := filepath.Join(GinkgoT().TempDir(), "bundle.tar")
			Expect(Copy(ctx, from, "archive:"+archivePath, copyOpts)).To(Succeed())
			Expect(Copy(ctx, "archive:"+archivePath, to, copyOpts)).To(Succeed())
		}),
	)

	It("should point the image values of werf-published dependencies at the new repo", func() {
		leafDir := writeTestChart()
		midDir := writeTestChart()
		Expect(os.MkdirAll(filepath.Join(midDir, "charts"), 0o755)).To(Succeed())
		packTestChart(ctx, leafDir, "leaf", filepath.Join(midDir, "charts"), werfDataChartPackOptions(ctx, fromRepo, imageDigest))

		rootDir := writeTestChart()
		Expect(os.MkdirAll(filepath.Join(rootDir, "charts"), 0o755)).To(Succeed())
		packTestChart(ctx, midDir, "mid", filepath.Join(rootDir, "charts"), werfDataChartPackOptions(ctx, fromRepo, imageDigest))

		Expect(Publish(ctx, rootDir, fromRepo, "1.2.3", testRegistryClient, PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())

		Expect(Copy(ctx, fromRepo+":1.2.3", toRepo+":4.5.6", copyOpts)).To(Succeed())

		Expect(testImageDigest(toRepo + ":tag-1")).To(Equal(imageDigest))

		mid := pullTestChart(ctx, testRegistryClient, toRepo, "4.5.6").Dependencies()[0]
		leaf := mid.Dependencies()[0]
		for _, dep := range []*chart.Chart{mid, leaf} {
			Expect(dep.Metadata.Annotations[chartyaml.ImagesAnnotation]).To(ContainSubstring(toRepo + ":tag-1@" + imageDigest))

			data, err := werfdata.Parse(dep.WerfDataFiles)
			Expect(err).To(Succeed())
			Expect(data.ExtraValues).To(Equal(werfDataExtraValues(toRepo, imageDigest)))
		}
	})

	DescribeTable("should keep the signature of a copy into a repo of the same name only when no werf values change",
		func(chartPackOpts func() action.ChartPackOptions, expectedErr string) {
			packOpts := chartPackOpts()
			packOpts.Sign = true
			packOpts.SignKey = "helm-test"
			packOpts.SignKeyring = testKeyringPath("helm-test-key.secret")
			Expect(Publish(ctx, writeTestChart(), fromRepo, "1.2.3", testRegistryClient, PublishOptions{ChartPackOptions: packOpts})).To(Succeed())

			sameNameRepo := path.Dir(toRepo) + "/" + path.Base(fromRepo)
			Expect(Copy(ctx, fromRepo+":1.2.3", sameNameRepo, copyOpts)).To(Succeed())

			err := Pull(ctx, sameNameRepo, "1.2.3", filepath.Join(GinkgoT().TempDir(), "bundle"), testRegistryClient, PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: ProvenanceStrategyVerify,
			})
			if expectedErr == "" {
				Expect(err).To(Succeed())
			} else {
				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
			}
		},
		Entry("no werf values", func() action.ChartPackOptions {
			return action.ChartPackOptions{Destination: GinkgoT().TempDir()}
		}, ""),
		Entry("werf values", func() action.ChartPackOptions {
			return werfDataChartPackOptions(ctx, fromRepo, imageDigest)
		}, "the bundle is not signed"),
	)

	DescribeTable("should keep a signed bundle with werf values unchanged when copied into the same repo under the same version",
		func(to func() string, tag string) {
			packOpts := werfDataChartPackOptions(ctx, fromRepo, imageDigest)
			packOpts.Sign = true
			packOpts.SignKey = "helm-test"
			packOpts.SignKeyring = testKeyringPath("helm-test-key.secret")
			Expect(Publish(ctx, writeTestChart(), fromRepo, "1.0.0", testRegistryClient, PublishOptions{ChartPackOptions: packOpts})).To(Succeed())
			source, err := testRegistryClient.Pull(fromRepo + ":1.0.0")
			Expect(err).To(Succeed())

			Expect(Copy(ctx, fromRepo+":1.0.0", to(), copyOpts)).To(Succeed())

			copied, err := testRegistryClient.Pull(fromRepo + ":" + tag)
			Expect(err).To(Succeed())
			Expect(copied.Manifest.Digest).To(Equal(source.Manifest.Digest))
			Expect(Pull(ctx, fromRepo, tag, filepath.Join(GinkgoT().TempDir(), "bundle"), testRegistryClient, PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: ProvenanceStrategyVerify,
			})).To(Succeed())
		},
		Entry("no tag", func() string { return fromRepo }, "1.0.0"),
		Entry("the tag of the same version", func() string { return fromRepo + ":v1.0.0" }, "v1.0.0"),
	)

	DescribeTable("should copy under the chart version and the tag",
		func(fromTag, toTag, expectedVersion string, expectedTags []string) {
			Expect(Publish(ctx, writeTestChart(), fromRepo, fromTag, testRegistryClient, PublishOptions{
				ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
			})).To(Succeed())

			to := toRepo
			if toTag != "" {
				to += ":" + toTag
			}

			Expect(Copy(ctx, fromRepo+":"+fromTag, to, copyOpts)).To(Succeed())

			repository, err := name.NewRepository(toRepo, name.Insecure)
			Expect(err).To(Succeed())
			tags, err := remote.List(repository)
			Expect(err).To(Succeed())
			Expect(tags).To(HaveLen(len(expectedTags)))
			for _, expectedTag := range expectedTags {
				Expect(tags).To(ContainElement(MatchRegexp(expectedTag)))
			}

			versionTag, _ := lo.Find(tags, func(tag string) bool { return regexp.MustCompile(expectedTags[0]).MatchString(tag) })
			Expect(pullTestChartVersion(ctx, testRegistryClient, toRepo, versionTag)).To(MatchRegexp(expectedVersion))
		},
		Entry("semver tag with the v prefix", "1.2.3", "v4.5.6", `^4\.5\.6$`, []string{`^4\.5\.6$`, `^v4\.5\.6$`}),
		Entry("non-semver tag", "1.2.3", "latest", `^0\.0\.0-\d+-latest$`, []string{`^0\.0\.0-\d+-latest$`, `^latest$`}),
		Entry("no tag keeps the source version and tag", "v1.2.3", "", `^1\.2\.3$`, []string{`^1\.2\.3$`, `^v1\.2\.3$`}),
		Entry("source tag with build metadata", "1.2.3+build.5", "", `^1\.2\.3\+build\.5$`, []string{`^1\.2\.3_build\.5$`}),
		Entry("destination tag with build metadata", "1.2.3", "4.5.6+build.5", `^4\.5\.6\+build\.5$`, []string{`^4\.5\.6_build\.5$`}),
	)

	It("should refuse a digest destination", func() {
		Expect(Copy(ctx, fromRepo+":tag-1", toRepo+"@"+imageDigest, copyOpts)).To(MatchError(ContainSubstring("digest")))
	})
})

func pushTestImage(ref string) string {
	img, err := random.Image(64, 1)
	Expect(err).To(Succeed())

	imageRef, err := name.ParseReference(ref, name.Insecure)
	Expect(err).To(Succeed())
	Expect(remote.Write(imageRef, img)).To(Succeed())

	digest, err := img.Digest()
	Expect(err).To(Succeed())

	return digest.String()
}

func testImageDigest(ref string) string {
	imageRef, err := name.ParseReference(ref, name.Insecure)
	Expect(err).To(Succeed())

	desc, err := remote.Head(imageRef)
	Expect(err).To(Succeed())

	return desc.Digest.String()
}

// oldBundleValues returns the values werf 3.12 and older publish the image and the repo in.
func oldBundleValues(repo, digest string) map[string]interface{} {
	imageVals, err := image.RebuildImageValuesMap(repo+":tag-1", digest)
	Expect(err).To(Succeed())

	return map[string]interface{}{
		"replicas": float64(1),
		"werf": map[string]interface{}{
			"env":   "production",
			"image": map[string]interface{}{"app": repo + ":tag-1"},
			"repo":  repo,
		},
		"global": map[string]interface{}{
			"werf": map[string]interface{}{
				"env":    "production",
				"repo":   repo,
				"images": map[string]interface{}{"app": imageVals},
			},
		},
	}
}

func werfDataExtraValues(repo, digest string) map[string]interface{} {
	return lo.OmitByKeys(oldBundleValues(repo, digest), []string{"replicas"})
}

func werfDataChartPackOptions(ctx context.Context, repo, digest string) action.ChartPackOptions {
	extraValues := werfDataExtraValues(repo, digest)

	return action.ChartPackOptions{
		Destination: GinkgoT().TempDir(),
		ExtraValues: extraValues,
		Images:      ChartImagesFromServiceValues(ctx, extraValues),
	}
}

// pushOldBundle pushes a bundle the way werf 3.12 and older publish it: its values hold the image and
// repo values and it has no .werf-data.
func pushOldBundle(client *registry.Client, repo, version string, values map[string]interface{}) {
	_, err := client.Push(oldBundleArchive(repo, version, values), repo+":"+version)
	Expect(err).To(Succeed())
}

func oldBundleArchive(repo, version string, values map[string]interface{}) []byte {
	chartDir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: "+path.Base(repo)+"\nversion: "+version+"\n"), 0o644)).To(Succeed())

	valuesYAML, err := yaml.Marshal(values)
	Expect(err).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chartDir, "values.yaml"), valuesYAML, 0o644)).To(Succeed())

	ch, err := loader.LoadDir(context.Background(), chartDir)
	Expect(err).To(Succeed())

	archivePath, err := chartv2util.Save(ch.(*chart.Chart), GinkgoT().TempDir())
	Expect(err).To(Succeed())

	archive, err := os.ReadFile(archivePath)
	Expect(err).To(Succeed())

	return archive
}

// writeWerf3Archive writes a bundle archive the way werf 3.12 and older write it: the chart archive
// and an image tarball per image tag, gzipped.
func writeWerf3Archive(path string, chartArchive []byte, images map[string]v1.Image) {
	f, err := os.Create(path)
	Expect(err).To(Succeed())
	defer f.Close()

	gzipWriter := gzip.NewWriter(f)
	tarWriter := tar.NewWriter(gzipWriter)

	writeMember := func(name string, data []byte) {
		Expect(tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg})).To(Succeed())
		_, err := tarWriter.Write(data)
		Expect(err).To(Succeed())
	}

	writeMember("chart.tar.gz", chartArchive)

	for tag, img := range images {
		var imageTarball bytes.Buffer
		imageGzipWriter := gzip.NewWriter(&imageTarball)
		imageRef, err := name.NewTag("image:" + tag)
		Expect(err).To(Succeed())
		Expect(tarball.Write(imageRef, img, imageGzipWriter)).To(Succeed())
		Expect(imageGzipWriter.Close()).To(Succeed())

		writeMember("images/"+tag+".tar.gz", imageTarball.Bytes())
	}

	Expect(tarWriter.Close()).To(Succeed())
	Expect(gzipWriter.Close()).To(Succeed())
}

func packTestChart(ctx context.Context, chartDir, name, destination string, opts action.ChartPackOptions) {
	opts.Destination = destination
	opts.Name = name
	opts.Version = "1.0.0"
	_, err := action.ChartPack(ctx, chartDir, opts)
	Expect(err).To(Succeed())
}

func pullTestChart(ctx context.Context, client *registry.Client, repo, tag string) *chart.Chart {
	bundleDir := filepath.Join(GinkgoT().TempDir(), "bundle")
	Expect(Pull(ctx, repo, tag, bundleDir, client, PullOptions{})).To(Succeed())

	ch, err := loader.Load(nelmcommon.ContextWithHelmOptions(ctx, nelmcommon.HelmOptions{
		ChartLoadOpts: nelmcommon.ChartLoadOptions{NoSecrets: true},
	}), bundleDir)
	Expect(err).To(Succeed())

	return ch.(*chart.Chart)
}
