package root

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
)

var _ = ginkgo.It("write-recording registry records a mutating request answered with an injected fault", func(ctx ginkgo.SpecContext) {
	repo, recorder := startRecordingRegistry("project/stages")
	recorder.fault.Store(&registryFault{pathPart: "/blobs/", status: http.StatusNotFound, code: "BLOB_UNKNOWN", message: "blob unknown to registry"})
	host, _, _ := strings.Cut(repo, "/")

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, "http://"+host+"/v2/project/stages/blobs/uploads/1", http.NoBody)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	response, err := http.DefaultClient.Do(request)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(response.Body.Close()).To(gomega.Succeed())

	gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusNotFound))
	gomega.Expect(recorder.recordedWrites()).To(gomega.ConsistOf("PUT /v2/project/stages/blobs/uploads/1"))
})

var _ = ginkgo.Describe("meta-repo info", func() {
	const (
		project  = "optional-docker"
		metaRepo = "registry.example/meta"
	)

	var (
		repo     string
		recorder *writeRecordingRegistry
		stages   *storage.RepoStagesStorage
		werfHome string
	)

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		repo, recorder = startRecordingRegistry("project/stages")
		stages = repoStagesStorageAt(ctx, repo)
		werfHome = ginkgo.GinkgoT().TempDir()
	})

	info := func(args ...string) (string, string, int) {
		ginkgo.GinkgoHelper()
		recorder.forgetWrites()
		host := "unix://" + filepath.Join(ginkgo.GinkgoT().TempDir(), "absent.sock")
		stdout, stderr, code := dockerCommandWithHome(dockerProject(false), host, werfHome, append([]string{"meta-repo", "info", "--repo", repo, "--insecure-registry"}, args...)...)
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty(), "stdout: %s\nstderr: %s", stdout, stderr)
		return stdout, stderr, code
	}

	ginkgo.It("reports a disabled safeguard for a repo without a marker", func() {
		stdout, stderr, code := info()

		gomega.Expect(code).To(gomega.BeZero(), "stderr: %s", stderr)
		gomega.Expect(stdout).To(gomega.Equal("Meta-repo path: not set\nSafeguard status: disabled\n"))
	})

	ginkgo.It("reports the meta-repo recorded in the marker", func(ctx ginkgo.SpecContext) {
		gomega.Expect(stages.PutMetaRepoMarker(ctx, project, metaRepo)).To(gomega.Succeed())

		stdout, stderr, code := info()

		gomega.Expect(code).To(gomega.BeZero(), "stderr: %s", stderr)
		gomega.Expect(stdout).To(gomega.Equal("Meta-repo path: " + metaRepo + "\nSafeguard status: enabled\n"))
	})

	ginkgo.DescribeTable("prints machine-readable output", func(ctx ginkgo.SpecContext, markerAddress string, safeguardEnabled bool) {
		if markerAddress != "" {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, project, markerAddress)).To(gomega.Succeed())
		}

		stdout, stderr, code := info("--output-format", "json")

		gomega.Expect(code).To(gomega.BeZero(), "stderr: %s", stderr)
		gomega.Expect(stdout).To(gomega.MatchJSON(fmt.Sprintf(`{"project":%q,"repo":%q,"safeguardEnabled":%t,"metaRepo":%q}`, project, repo, safeguardEnabled, markerAddress)))
	},
		ginkgo.Entry("without a marker", "", false),
		ginkgo.Entry("with a marker", metaRepo, true),
	)

	ginkgo.It("keeps machine-readable output parseable while a host cleanup notice is pending", func() {
		noticePath := filepath.Join(werfHome, "service", "auto_host_cleanup_notice.json")
		gomega.Expect(os.MkdirAll(filepath.Dir(noticePath), 0o755)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(noticePath, []byte(`{"backend_name":"docker","space_reclaimed":1}`), 0o644)).To(gomega.Succeed())

		stdout, stderr, code := info("--output-format", "json")

		gomega.Expect(code).To(gomega.BeZero(), "stderr: %s", stderr)
		gomega.Expect(stdout).To(gomega.MatchJSON(fmt.Sprintf(`{"project":%q,"repo":%q,"safeguardEnabled":false,"metaRepo":""}`, project, repo)))
		gomega.Expect(stderr).To(gomega.ContainSubstring("Automatic host cleanup ran in background"))
	})

	ginkgo.DescribeTable("keeps machine-readable output parseable with diagnostic logs enabled", func(ctx ginkgo.SpecContext, logFlag string) {
		gomega.Expect(stages.PutMetaRepoMarker(ctx, project, metaRepo)).To(gomega.Succeed())

		stdout, stderr, code := info("--output-format", "json", logFlag)

		gomega.Expect(code).To(gomega.BeZero(), "stderr: %s", stderr)
		gomega.Expect(stdout).To(gomega.MatchJSON(fmt.Sprintf(`{"project":%q,"repo":%q,"safeguardEnabled":true,"metaRepo":%q}`, project, repo, metaRepo)))
		gomega.Expect(stderr).NotTo(gomega.BeEmpty(), "diagnostic logs must go to stderr")
	},
		ginkgo.Entry("verbose", "--log-verbose"),
		ginkgo.Entry("debug", "--log-debug"),
	)

	markerRef := func() string {
		return fmt.Sprintf(storage.RepoMetaRepoMarker_ImageNameFormat, repo, project)
	}

	putMarker := func(ctx context.Context) {
		ginkgo.GinkgoHelper()
		gomega.Expect(stages.PutMetaRepoMarker(ctx, project, metaRepo)).To(gomega.Succeed())
	}

	putIndexAtMarkerTag := func(ctx context.Context) {
		ginkgo.GinkgoHelper()
		ref, err := name.ParseReference(markerRef(), name.Insecure)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		index := mutate.AppendManifests(empty.Index, mutate.IndexAddendum{Add: empty.Image})
		gomega.Expect(remote.WriteIndex(ref, index, remote.WithContext(ctx))).To(gomega.Succeed())
	}

	blobMissing := registryFault{pathPart: "/blobs/", status: http.StatusNotFound, code: "BLOB_UNKNOWN", message: "blob unknown to registry"}
	childManifestMissing := registryFault{pathPart: "/manifests/sha256:", status: http.StatusNotFound, code: "MANIFEST_UNKNOWN", message: "manifest unknown"}

	ginkgo.DescribeTable("fails when the marker tag exists but cannot be read", func(ctx ginkgo.SpecContext, plant func(context.Context), fault registryFault, outputFormat string) {
		plant(ctx)
		recorder.fault.Store(&fault)

		stdout, stderr, code := info("--output-format", outputFormat)

		gomega.Expect(code).To(gomega.Equal(1), "stdout: %s", stdout)
		gomega.Expect(stdout).To(gomega.BeEmpty())
		gomega.Expect(stderr).To(gomega.And(gomega.ContainSubstring("werf-meta-repo-marker-"+project), gomega.ContainSubstring(fault.code)))
	},
		ginkgo.Entry("config blob is missing, text", putMarker, blobMissing, "text"),
		ginkgo.Entry("config blob is missing, json", putMarker, blobMissing, "json"),
		ginkgo.Entry("child manifest of an index is missing, text", putIndexAtMarkerTag, childManifestMissing, "text"),
		ginkgo.Entry("child manifest of an index is missing, json", putIndexAtMarkerTag, childManifestMissing, "json"),
	)

	ginkgo.DescribeTable("fails when the registry does not plainly answer that the marker tag is missing", func(status int, errorCode, message, outputFormat string) {
		recorder.fault.Store(&registryFault{pathPart: "/manifests/werf-meta-repo-marker-", status: status, code: errorCode, message: message})

		stdout, stderr, code := info("--output-format", outputFormat)

		gomega.Expect(code).To(gomega.Equal(1), "stdout: %s", stdout)
		gomega.Expect(stdout).To(gomega.BeEmpty())
		gomega.Expect(stderr).To(gomega.And(gomega.ContainSubstring("werf-meta-repo-marker-"+project), gomega.ContainSubstring(errorCode)))
	},
		ginkgo.Entry("access denied", http.StatusForbidden, "DENIED", "read access denied", "json"),
		ginkgo.Entry("access denied with a not-found code in the message, text", http.StatusForbidden, "DENIED", "permission lookup failed: NOT_FOUND", "text"),
		ginkgo.Entry("access denied with a not-found code in the message, json", http.StatusForbidden, "DENIED", "permission lookup failed: NOT_FOUND", "json"),
		ginkgo.Entry("not-found status with another code", http.StatusNotFound, "DENIED", "upstream returned MANIFEST_UNKNOWN", "json"),
		ginkgo.Entry("registry failure with a not-found code", http.StatusInternalServerError, "MANIFEST_UNKNOWN", "upstream registry failure", "json"),
	)

	ginkgo.It("fails on a marker without a meta-repo address and points at detach", func(ctx ginkgo.SpecContext) {
		gomega.Expect(stages.PutMetaRepoMarker(ctx, project, "")).To(gomega.Succeed())

		stdout, stderr, code := info()

		gomega.Expect(code).To(gomega.Equal(1), "stdout: %s", stdout)
		gomega.Expect(stdout).To(gomega.BeEmpty())
		gomega.Expect(stderr).To(gomega.And(gomega.ContainSubstring("malformed"), gomega.ContainSubstring("werf meta-repo detach")))
	})

	ginkgo.It("fails when the reserved tag is occupied by a stage image", func(ctx ginkgo.SpecContext) {
		gomega.Expect(stages.DockerRegistry.PushImage(ctx, markerRef(), &docker_registry.PushImageOptions{
			Labels: map[string]string{image.WerfStageContentDigestLabel: "digest"},
		})).To(gomega.Succeed())

		stdout, stderr, code := info()

		gomega.Expect(code).To(gomega.Equal(1), "stdout: %s", stdout)
		gomega.Expect(stdout).To(gomega.BeEmpty())
		gomega.Expect(stderr).To(gomega.ContainSubstring("werf reserves that tag for the meta-repo safeguard"))
	})

	ginkgo.It("rejects an unknown output format", func() {
		stdout, stderr, code := info("--output-format", "yaml")

		gomega.Expect(code).To(gomega.Equal(1), "stdout: %s", stdout)
		gomega.Expect(stderr).To(gomega.ContainSubstring(`--output-format must be "text" or "json"`))
	})
})
