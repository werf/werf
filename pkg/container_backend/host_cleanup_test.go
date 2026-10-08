package container_backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("host cleanup service", func() {
	ginkgo.BeforeEach(func() {
		for _, name := range []string{"WERF_HOST_CLEANUP_SERVICE_IMAGE", "WERF_STAPEL_IMAGE_NAME", "WERF_STAPEL_IMAGE_VERSION"} {
			ginkgo.GinkgoT().Setenv(name, "")
		}
		root := ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("WERF_HOME", filepath.Join(root, "home"))
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", root)
		gomega.Expect(werf.Init(root, filepath.Join(root, "home"))).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("uses the selected image and command in Docker", func(override, stapelName, stapelVersion, expectedImage, expectedCommand string) {
		const mountDir = `/tmp/selected, "quoted"`
		ginkgo.GinkgoT().Setenv("WERF_HOST_CLEANUP_SERVICE_IMAGE", override)
		ginkgo.GinkgoT().Setenv("WERF_STAPEL_IMAGE_NAME", stapelName)
		ginkgo.GinkgoT().Setenv("WERF_STAPEL_IMAGE_VERSION", stapelVersion)
		type containerRequest struct {
			Image      string
			Cmd        []string
			HostConfig struct {
				Mounts []struct{ Type, Source, Target string }
			}
		}
		requests := make(chan containerRequest, 1)
		var inspected atomic.Bool
		ctx, _ := dockerDaemonContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/_ping"):
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
			case strings.Contains(r.URL.Path, "/images/") && strings.HasSuffix(r.URL.Path, "/json"):
				gomega.Expect(override).To(gomega.BeEmpty())
				gomega.Expect(r.URL.Path).To(gomega.ContainSubstring(expectedImage))
				inspected.Store(true)
				_, err := io.WriteString(w, `{"Id":"sha256:1234","Os":"linux","Architecture":"amd64","Config":{}}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			case strings.HasSuffix(r.URL.Path, "/containers/create"):
				var request containerRequest
				gomega.Expect(json.NewDecoder(r.Body).Decode(&request)).To(gomega.Succeed())
				requests <- request
				w.WriteHeader(http.StatusInternalServerError)
				_, err := io.WriteString(w, `{"message":"cleanup create error"}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			default:
				ginkgo.Fail(fmt.Sprintf("unexpected Docker request: %s %s", r.Method, r.URL.Path))
			}
		}))
		err := (&DockerServerBackend{}).RemoveHostDirs(ctx, mountDir, []string{mountDir + "/cache"})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("cleanup create error")))
		gomega.Expect(inspected.Load()).To(gomega.Equal(override == ""))
		var request containerRequest
		gomega.Expect(requests).To(gomega.Receive(&request))
		gomega.Expect(request.Image).To(gomega.Equal(expectedImage))
		gomega.Expect(request.Cmd).To(gomega.Equal([]string{expectedCommand, "-rf", "--", mountDir + "/cache"}))
		gomega.Expect(request.HostConfig.Mounts).To(gomega.HaveLen(1))
		gomega.Expect(request.HostConfig.Mounts[0].Type).To(gomega.Equal("bind"))
		gomega.Expect(request.HostConfig.Mounts[0].Source).To(gomega.Equal(mountDir))
		gomega.Expect(request.HostConfig.Mounts[0].Target).To(gomega.Equal(mountDir))
	},
		ginkgo.Entry("default Stapel", "", "", "", "registry.werf.io/werf/stapel:"+stapel.VERSION, "/.werf/stapel/embedded/bin/rm"),
		ginkgo.Entry("configured Stapel mirror", "", "mirror.example/stapel", "custom", "mirror.example/stapel:custom", "/.werf/stapel/embedded/bin/rm"),
		ginkgo.Entry("explicit cleanup image wins", "custom.example/cleanup:1", "mirror.example/stapel", "custom", "custom.example/cleanup:1", "rm"),
	)

	ginkgo.It("reports a Stapel preparation error before starting a cleanup container", func() {
		ctx, _ := dockerDaemonContext(daemonHandler(http.StatusInternalServerError, `{"message":"stapel inspect error"}`))
		err := (&DockerServerBackend{}).RemoveHostDirs(ctx, "/tmp/selected", []string{"/tmp/selected/cache"})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("prepare stapel for host cleanup")))
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("stapel inspect error")))
	})

	ginkgo.DescribeTable("uses the selected image and command in Buildah", func(override, expectedImage, expectedCommand string) {
		ginkgo.GinkgoT().Setenv("WERF_HOST_CLEANUP_SERVICE_IMAGE", override)
		runError := errors.New("cleanup command error")
		runCalled := false
		ctx := context.Background()
		stub := &hostCleanupBuildah{runCommand: func(commandCtx context.Context, container string, command []string, opts buildah.RunCommandOpts) error {
			runCalled = true
			gomega.Expect(commandCtx).To(gomega.BeIdenticalTo(ctx))
			gomega.Expect(container).NotTo(gomega.BeEmpty())
			gomega.Expect(command).To(gomega.Equal([]string{expectedCommand, "-rf", "--", "/tmp/selected/cache"}))
			gomega.Expect(opts.User).To(gomega.Equal("0:0"))
			gomega.Expect(opts.GlobalMounts).To(gomega.HaveLen(1))
			gomega.Expect(opts.GlobalMounts[0].Source).To(gomega.Equal("/tmp/selected"))
			gomega.Expect(opts.GlobalMounts[0].Destination).To(gomega.Equal("/tmp/selected"))
			return runError
		}}
		backend := NewBuildahBackend(stub, BuildahBackendOptions{})
		gomega.Expect(backend.RemoveHostDirs(ctx, "/tmp/selected", []string{"/tmp/selected/cache"})).To(gomega.MatchError(runError))
		gomega.Expect(runCalled).To(gomega.BeTrue())
		gomega.Expect(stub.FromCommandImages).To(gomega.Equal([]string{expectedImage}))
	},
		ginkgo.Entry("default Stapel", "", "registry.werf.io/werf/stapel:"+stapel.VERSION, "/.werf/stapel/embedded/bin/rm"),
		ginkgo.Entry("explicit cleanup image", "custom.example/cleanup:1", "custom.example/cleanup:1", "rm"),
	)
})
