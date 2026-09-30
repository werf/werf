package common

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/moby/moby/client"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker"
)

var _ = ginkgo.DescribeTable("InitProcessDocker checks the daemon API only when the command needs the daemon", func(apiVersion, expectedError string, alreadyBound, requireDaemon bool) {
	oldDockerConfigDir, oldCLIConfigDir := docker.DockerConfigDir, cliconfig.Dir()
	ginkgo.DeferCleanup(func() {
		docker.DockerConfigDir = oldDockerConfigDir
		cliconfig.SetDir(oldCLIConfigDir)
	})
	for _, key := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION", "DOCKER_DEFAULT_PLATFORM", "WERF_DEBUG_DOCKER"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	var advertised atomic.Value
	advertised.Store(apiVersion)
	if alreadyBound {
		advertised.Store("1.40")
	}
	var pings, versionedRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		if r.URL.Path == "/_ping" {
			pings.Add(1)
			w.Header().Set("API-Version", advertised.Load().(string))
			w.Header().Set("OSType", "linux")
			return
		}
		versionedRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`[]`))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}))
	ginkgo.DeferCleanup(server.Close)
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	dockerConfigDir := ginkgo.GinkgoT().TempDir()
	verbose, debug := false, false
	cmdData := &CmdData{DockerConfig: &dockerConfigDir, LogVerbose: &verbose, LogDebug: &debug}
	ctx := context.Background()
	if alreadyBound {
		gomega.Expect(docker.InitDockerConfig(docker.InitOptions{DockerConfigDir: dockerConfigDir})).To(gomega.Succeed())
		var err error
		ctx, err = docker.NewContext(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = docker.Images(ctx, client.ImageListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(versionedRequests.Load()).To(gomega.Equal(int32(1)))
		advertised.Store(apiVersion)
	}
	before := versionedRequests.Load()
	if apiVersion == "unavailable" {
		server.Close()
	}
	boundCtx, err := InitProcessDocker(ctx, cmdData, InitProcessDockerOptions{RequireDaemon: requireDaemon})
	if apiVersion != "unavailable" {
		gomega.Expect(pings.Load()).NotTo(gomega.BeZero())
	}
	gomega.Expect(versionedRequests.Load()).To(gomega.Equal(before))
	if expectedError != "" {
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(expectedError))
		gomega.Expect(docker.IsContext(boundCtx)).To(gomega.Equal(alreadyBound))
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(docker.IsContext(boundCtx)).To(gomega.BeTrue())
},
	ginkgo.Entry("supported API version", "1.40", "", false, true),
	ginkgo.Entry("unsupported API version", "1.39", "minimum supported API version is 1.40", false, true),
	ginkgo.Entry("unsupported API version is tolerated by registry-only commands", "1.39", "", false, false),
	ginkgo.Entry("missing API version header", "", "did not report an API version", false, true),
	ginkgo.Entry("missing API version header is tolerated by registry-only commands", "", "", false, false),
	ginkgo.Entry("already bound context cannot bypass the guard", "1.39", "minimum supported API version is 1.40", true, true),
	ginkgo.Entry("already bound context is tolerated by registry-only commands", "1.39", "", true, false),
	ginkgo.Entry("unavailable daemon remains optional", "unavailable", "", false, true),
	ginkgo.Entry("unavailable daemon remains optional for bound context", "unavailable", "", true, true),
)
