package stapel

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("host cleanup Stapel image", func() {
	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("DOCKER_CONTEXT", "unused-stapel-test-context")
		for _, key := range []string{"WERF_STAPEL_IMAGE_NAME", "WERF_STAPEL_IMAGE_VERSION"} {
			ginkgo.GinkgoT().Setenv(key, "")
		}
		root := ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("WERF_HOME", filepath.Join(root, "home"))
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", root)
		gomega.Expect(werf.Init(root, filepath.Join(root, "home"))).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("prepares the cleanup image for the requested platform", func(arch string) {
		ctx := stapelDaemonContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/_ping"):
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
			case strings.Contains(r.URL.Path, "/images/") && strings.HasSuffix(r.URL.Path, "/json"):
				gomega.Expect(r.URL.Path).To(gomega.ContainSubstring("registry.werf.io/werf/stapel:0.7.2"))
				_, err := fmt.Fprintf(w, `{"Id":"sha256:1234","Os":"linux","Architecture":%q,"Config":{}}`, arch)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			default:
				ginkgo.Fail(fmt.Sprintf("unexpected Docker request: %s %s", r.Method, r.URL.Path))
			}
		}))
		gomega.Expect(EnsureImage(ctx, HostCleanupImageName(ctx), "linux/"+arch)).To(gomega.Succeed())
		gomega.Expect(ImageName()).To(gomega.Equal("registry.werf.io/werf/stapel:0.7.2"))
	}, ginkgo.Entry("amd64", "amd64"), ginkgo.Entry("arm64", "arm64"))

	ginkgo.DescribeTable("purges build and cleanup images without duplicating deletion", func(name, version string, expected []string) {
		ginkgo.GinkgoT().Setenv("WERF_STAPEL_IMAGE_NAME", name)
		ginkgo.GinkgoT().Setenv("WERF_STAPEL_IMAGE_VERSION", version)
		images := lo.SliceToMap(expected, func(imageName string) (string, bool) { return imageName, true })
		var deleted []string
		var mu sync.Mutex
		ctx := stapelDaemonContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			mu.Lock()
			defer mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/_ping"):
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
			case strings.HasSuffix(r.URL.Path, "/containers/json"):
				_, err := io.WriteString(w, "[]")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			case strings.Contains(r.URL.Path, "/images/"):
				_, imageName, found := strings.Cut(r.URL.Path, "/images/")
				gomega.Expect(found).To(gomega.BeTrue())
				imageName = strings.TrimSuffix(imageName, "/json")
				if !images[imageName] {
					w.WriteHeader(http.StatusNotFound)
					_, err := io.WriteString(w, `{"message":"image missing"}`)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				if r.Method == http.MethodDelete {
					deleted = append(deleted, imageName)
					delete(images, imageName)
					_, err := io.WriteString(w, "[]")
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
				} else {
					_, err := io.WriteString(w, `{"Id":"sha256:1234","Os":"linux","Architecture":"amd64","Config":{}}`)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
				}
			default:
				ginkgo.Fail(fmt.Sprintf("unexpected Docker request: %s %s", r.Method, r.URL.Path))
			}
		}))
		gomega.Expect(Purge(ctx)).To(gomega.Succeed())
		mu.Lock()
		defer mu.Unlock()
		gomega.Expect(images).To(gomega.BeEmpty())
		gomega.Expect(deleted).To(gomega.ConsistOf(expected))
	},
		ginkgo.Entry("default images", "", "", []string{"registry.werf.io/werf/stapel:0.7.2"}),
		ginkgo.Entry("configured image", "mirror.example/stapel", "custom", []string{"mirror.example/stapel:custom"}),
	)
})
