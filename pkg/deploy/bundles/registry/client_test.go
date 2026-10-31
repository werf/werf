package registry

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	ggcrregistry "github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/werf/v3/pkg/ref"
)

var _ = Describe("Client", func() {
	var sourceRegistryHost, destinationRegistryHost string

	BeforeEach(func() {
		sourceRegistryHost = startTestRegistry()
		destinationRegistryHost = startTestRegistry()
	})

	It("should keep the provenance layer of a signed chart through pull, cache and push", func(ctx context.Context) {
		provenance := []byte("-----BEGIN PGP SIGNED MESSAGE-----\nprovenance\n")
		sourceRef := &ref.Reference{Repo: sourceRegistryHost + "/group/signedchart", Tag: "0.1.0"}

		publisher := newTestClient(ctx)
		storeSignedChart(ctx, publisher.cache, sourceRef, provenance)
		Expect(publisher.PushChart(ctx, sourceRef, []string{sourceRef.Tag}, nelmcommon.HelmOptions{})).To(Succeed())

		client := newTestClient(ctx)
		Expect(client.PullChartToCache(ctx, sourceRef, nelmcommon.HelmOptions{})).To(Succeed())

		signed, err := client.IsChartSigned(ctx, sourceRef, nelmcommon.HelmOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(signed).To(BeTrue())

		ch, err := client.LoadChart(ctx, sourceRef, nelmcommon.HelmOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(ch.Metadata.Name).To(Equal("signedchart"))

		destinationRef := &ref.Reference{Repo: destinationRegistryHost + "/group2/signedchart", Tag: "0.1.0"}
		Expect(client.TagCachedChart(ctx, sourceRef, destinationRef, nelmcommon.HelmOptions{})).To(Succeed())
		Expect(client.PushChart(ctx, destinationRef, []string{"0.1.0", "v0.1.0"}, nelmcommon.HelmOptions{})).To(Succeed())

		for _, tag := range []string{"0.1.0", "v0.1.0"} {
			layers := remoteManifestLayers(destinationRegistryHost + "/group2/signedchart:" + tag)
			Expect(layers).To(HaveLen(2))
			Expect(layers[0].MediaType).To(Equal(HelmChartContentLayerFullMediaType))
			Expect(layers[1].MediaType).To(Equal(HelmChartProvenanceLayerMediaType))
			Expect(layers[1].Digest).To(Equal(digest.FromBytes(provenance)))
		}
	})
})

func startTestRegistry() string {
	server := httptest.NewServer(ggcrregistry.New(ggcrregistry.Logger(log.New(io.Discard, "", 0))))
	DeferCleanup(server.Close)

	return strings.TrimPrefix(server.URL, "http://")
}

func newTestClient(ctx context.Context) *Client {
	tmpDir := GinkgoT().TempDir()

	cache, err := NewCache(CacheOptRoot(filepath.Join(tmpDir, "cache")))
	Expect(err).NotTo(HaveOccurred())

	client, err := NewClient(ctx, ClientOptCredentialsFile(filepath.Join(tmpDir, "config.json")), ClientOptInsecure(true), ClientOptCache(cache))
	Expect(err).NotTo(HaveOccurred())

	return client
}

func storeSignedChart(ctx context.Context, cache *Cache, chartRef *ref.Reference, provenance []byte) {
	Expect(cache.init()).To(Succeed())

	ch := &chart.Chart{Metadata: &chart.Metadata{APIVersion: "v2", Name: "signedchart", Version: "0.1.0"}}

	archivePath, err := chartv2util.Save(ch, GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	archive, err := os.ReadFile(archivePath)
	Expect(err).NotTo(HaveOccurred())

	config, err := json.Marshal(ch.Metadata)
	Expect(err).NotTo(HaveOccurred())

	manifest, err := json.Marshal(ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config:    storeTestBlob(ctx, cache, HelmChartConfigMediaType, config),
		Layers: []ocispec.Descriptor{
			storeTestBlob(ctx, cache, HelmChartContentLayerFullMediaType, archive),
			storeTestBlob(ctx, cache, HelmChartProvenanceLayerMediaType, provenance),
		},
	})
	Expect(err).NotTo(HaveOccurred())

	manifestDesc := storeTestBlob(ctx, cache, ocispec.MediaTypeImageManifest, manifest)
	Expect(cache.AddManifest(ctx, chartRef, &manifestDesc)).To(Succeed())
}

func storeTestBlob(ctx context.Context, cache *Cache, mediaType string, data []byte) ocispec.Descriptor {
	_, err := cache.storeBlob(ctx, data)
	Expect(err).NotTo(HaveOccurred())

	return ocispec.Descriptor{MediaType: mediaType, Digest: digest.FromBytes(data), Size: int64(len(data))}
}

func remoteManifestLayers(reference string) []ocispec.Descriptor {
	parsedRef, err := name.ParseReference(reference, name.Insecure)
	Expect(err).NotTo(HaveOccurred())

	desc, err := remote.Get(parsedRef)
	Expect(err).NotTo(HaveOccurred())

	var manifest ocispec.Manifest
	Expect(json.Unmarshal(desc.Manifest, &manifest)).To(Succeed())

	return manifest.Layers
}
