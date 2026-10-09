//go:build embedstapel

package stapel

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("embedded host cleanup image", func() {
	ginkgo.It("loads the missing default image without pulling from a registry", func() {
		for _, key := range []string{"WERF_STAPEL_IMAGE_NAME", "WERF_STAPEL_IMAGE_VERSION"} {
			ginkgo.GinkgoT().Setenv(key, "")
		}
		root := ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("WERF_HOME", filepath.Join(root, "home"))
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", root)
		gomega.Expect(werf.Init(root, filepath.Join(root, "home"))).To(gomega.Succeed())
		var mu sync.Mutex
		var loadedBytes int64
		var pulled bool
		ctx := stapelDaemonContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			mu.Lock()
			defer mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/_ping"):
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
			case strings.Contains(r.URL.Path, "/images/") && strings.HasSuffix(r.URL.Path, "/json"):
				if loadedBytes == 0 {
					w.WriteHeader(http.StatusNotFound)
					_, err := io.WriteString(w, `{"message":"image missing"}`)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				_, err := fmt.Fprintf(w, `{"Id":"sha256:loaded","Os":"linux","Architecture":%q,"Config":{}}`, runtime.GOARCH)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			case strings.HasSuffix(r.URL.Path, "/images/load"):
				n, err := io.Copy(io.Discard, r.Body)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				loadedBytes = n
				_, err = io.WriteString(w, `{"stream":"Loaded image\n"}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			case strings.HasSuffix(r.URL.Path, "/images/create"):
				pulled = true
				_, err := io.WriteString(w, `{"status":"pulled"}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			default:
				ginkgo.Fail(fmt.Sprintf("unexpected Docker request: %s %s", r.Method, r.URL.Path))
			}
		}))
		gomega.Expect(EnsureImage(ctx, HostCleanupImageName(ctx), "linux/"+runtime.GOARCH)).To(gomega.Succeed())
		mu.Lock()
		defer mu.Unlock()
		gomega.Expect(loadedBytes).To(gomega.BeNumerically(">", 0))
		gomega.Expect(pulled).To(gomega.BeFalse())
	})
})
