package bundles

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

	_, err = client.Push(archive, repo+":"+version)
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
