package bundles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("Bundle pull", func() {
	DescribeTable("should trust only the chart archive, not the chart metadata in the OCI config",
		func(ctx0 context.Context, config, strategy string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"

			packed, err := action.ChartPack(ctx, writeTestChart(), signingChartPackOptions())
			Expect(err).To(Succeed())
			chartArchive, err := os.ReadFile(packed.ArchivePath)
			Expect(err).To(Succeed())
			provenanceData, err := os.ReadFile(packed.ProvenancePath)
			Expect(err).To(Succeed())
			pushRawChart(repo, "1.0.0", []byte(config), chartArchive, provenanceData)

			baseDir := GinkgoT().TempDir()
			victimFile := filepath.Join(baseDir, "victim", "file")
			Expect(os.MkdirAll(filepath.Dir(victimFile), 0o755)).To(Succeed())
			Expect(os.WriteFile(victimFile, []byte("victim"), 0o644)).To(Succeed())

			bundleDir := filepath.Join(baseDir, "bundle")
			Expect(Pull(ctx, repo, "1.0.0", bundleDir, newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: strategy,
			})).To(Succeed())

			Expect(filepath.Join(bundleDir, "Chart.yaml")).To(BeARegularFile())
			Expect(os.ReadFile(victimFile)).To(Equal([]byte("victim")))
		},
		Entry("a name escaping the directory, verify", `{"apiVersion":"v2","name":"../victim","version":"9.9.9"}`, ProvenanceStrategyVerify),
		Entry("a name escaping the directory, never verify", `{"apiVersion":"v2","name":"../victim","version":"9.9.9"}`, "never"),
		Entry("null, verify", `null`, ProvenanceStrategyVerify),
		Entry("null, never verify", `null`, "never"),
	)
})

// pushRawChart pushes a chart with the OCI config as given, which the registry client would generate
// from the chart.
func pushRawChart(repo, tag string, config, chartArchive, provenanceData []byte) {
	host, repoPath, _ := strings.Cut(repo, "/")
	baseURL := fmt.Sprintf("http://%s/v2/%s", host, repoPath)

	pushBlob := func(data []byte, mediaType string) map[string]interface{} {
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(data))

		resp, err := http.Post(fmt.Sprintf("%s/blobs/uploads/?digest=%s", baseURL, digest), "application/octet-stream", bytes.NewReader(data))
		Expect(err).To(Succeed())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		return map[string]interface{}{"mediaType": mediaType, "digest": digest, "size": len(data)}
	}

	manifest, err := json.Marshal(map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config":        pushBlob(config, registry.ConfigMediaType),
		"layers": []interface{}{
			pushBlob(chartArchive, registry.ChartLayerMediaType),
			pushBlob(provenanceData, registry.ProvLayerMediaType),
		},
	})
	Expect(err).To(Succeed())

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/manifests/%s", baseURL, tag), bytes.NewReader(manifest))
	Expect(err).To(Succeed())
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	resp, err := http.DefaultClient.Do(req)
	Expect(err).To(Succeed())
	Expect(resp.Body.Close()).To(Succeed())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
}
