package docker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = ginkgo.DescribeTable("docker api operations", func(ctx ginkgo.SpecContext, label opstats.Operation, handler http.HandlerFunc, call func(ctx context.Context) error, expectError bool) {
	apiCtx, collector := observedDaemonContext(handler)

	err := call(apiCtx)
	if expectError {
		gomega.Expect(err).To(gomega.HaveOccurred())
	} else {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	gomega.Expect(collector.Summary()).To(gomega.HaveLen(1), "exactly one operation must be recorded")
	gomega.Expect(operationCount(collector, label)).To(gomega.Equal(1))
},
	ginkgo.Entry("image list", opstats.Operation("docker: image list"), respondJSON(`[]`), func(ctx context.Context) error {
		_, err := Images(ctx, types.ImageListOptions{})
		return err
	}, false),
	ginkgo.Entry("image list failure", opstats.Operation("docker: image list"), http.HandlerFunc(respondError), func(ctx context.Context) error {
		_, err := Images(ctx, types.ImageListOptions{})
		return err
	}, true),
	ginkgo.Entry("image inspect", opstats.Operation("docker: image inspect"), respondJSON(`{"Id":"sha256:abc"}`), func(ctx context.Context) error {
		_, err := ImageInspect(ctx, "img")
		return err
	}, false),
	ginkgo.Entry("image inspect failure", opstats.Operation("docker: image inspect"), http.HandlerFunc(respondError), func(ctx context.Context) error {
		_, err := ImageInspect(ctx, "img")
		return err
	}, true),
	ginkgo.Entry("container list", opstats.Operation("docker: container list"), respondJSON(`[]`), func(ctx context.Context) error {
		_, err := Containers(ctx, types.ContainerListOptions{})
		return err
	}, false),
	ginkgo.Entry("container inspect", opstats.Operation("docker: container inspect"), respondJSON(`{"Id":"c"}`), func(ctx context.Context) error {
		_, err := ContainerInspect(ctx, "c")
		return err
	}, false),
	ginkgo.Entry("container commit", opstats.Operation("docker: container commit"), respondJSON(`{"Id":"sha256:abc"}`), func(ctx context.Context) error {
		_, err := ContainerCommit(ctx, "c", types.ContainerCommitOptions{})
		return err
	}, false),
	ginkgo.Entry("container commit failure", opstats.Operation("docker: container commit"), http.HandlerFunc(respondError), func(ctx context.Context) error {
		_, err := ContainerCommit(ctx, "c", types.ContainerCommitOptions{})
		return err
	}, true),
	ginkgo.Entry("container remove", opstats.Operation("docker: container remove"), respondJSON(``), func(ctx context.Context) error {
		return ContainerRemove(ctx, "c", types.ContainerRemoveOptions{})
	}, false),
	ginkgo.Entry("image prune", opstats.Operation("docker: image prune"), respondJSON(`{"ImagesDeleted":[],"SpaceReclaimed":0}`), func(ctx context.Context) error {
		_, err := ImagesPrune(ctx, ImagesPruneOptions{})
		return err
	}, false),
	ginkgo.Entry("image load", opstats.Operation("docker: image load"), respondJSON(`{"stream":"Loaded image ID: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"}`), func(ctx context.Context) error {
		_, err := CliLoadFromStream(ctx, strings.NewReader(""))
		return err
	}, false),
)

var _ = ginkgo.Describe("docker container create", func() {
	ginkgo.It("records a single container create", func(ctx ginkgo.SpecContext) {
		apiCtx, collector := observedDaemonContext(respondJSON(`{"Id":"c"}`))

		err := CliCreate(apiCtx, "img")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(collector.Summary()).To(gomega.HaveLen(1))
		gomega.Expect(operationCount(collector, "docker: container create")).To(gomega.Equal(1))
	})
})

var _ = ginkgo.DescribeTable("docker stream operations", func(ctx ginkgo.SpecContext, label opstats.Operation, open func(ctx context.Context) (io.ReadCloser, error)) {
	apiCtx, collector := observedDaemonContext(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		w.Header().Set("Content-Type", "application/x-tar")
		if _, err := io.WriteString(w, "stream payload"); err != nil {
			panic(err)
		}
	})

	rc, err := open(apiCtx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(collector.Summary()).To(gomega.BeEmpty(), "the operation must stay in flight until the stream is consumed")

	data, err := io.ReadAll(rc)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(data)).To(gomega.Equal("stream payload"))
	gomega.Expect(rc.Close()).To(gomega.Succeed())

	gomega.Expect(collector.Summary()).To(gomega.HaveLen(1))
	gomega.Expect(operationCount(collector, label)).To(gomega.Equal(1), "reading to EOF and closing must record the operation once")
},
	ginkgo.Entry("image save", opstats.Operation("docker: image save"), func(ctx context.Context) (io.ReadCloser, error) {
		return CliImageSaveToStream(ctx, "img")
	}),
)

var _ = ginkgo.DescribeTable("docker stream operations failing before the stream", func(ctx ginkgo.SpecContext, label opstats.Operation, open func(ctx context.Context) (io.ReadCloser, error)) {
	apiCtx, collector := observedDaemonContext(http.HandlerFunc(respondError))

	_, err := open(apiCtx)
	gomega.Expect(err).To(gomega.HaveOccurred())

	gomega.Expect(operationCount(collector, label)).To(gomega.Equal(1), "a failure before the stream must still record the operation")
},
	ginkgo.Entry("image save", opstats.Operation("docker: image save"), func(ctx context.Context) (io.ReadCloser, error) {
		return CliImageSaveToStream(ctx, "img")
	}),
)

var _ = ginkgo.DescribeTable("docker cli registry transfers", func(ctx ginkgo.SpecContext, label opstats.Operation, call func(ctx context.Context) error) {
	cliCtx := cliDaemonContext(newMissingImageDaemonServer())

	collector := opstats.NewCollector()
	gomega.Expect(call(opstats.NewContext(cliCtx, collector))).NotTo(gomega.Succeed())

	summary := collector.Summary()
	gomega.Expect(summary).To(gomega.HaveLen(1), "exactly one operation must be recorded")
	gomega.Expect(summary[0].Operation).To(gomega.Equal(label))
	gomega.Expect(summary[0].Count).To(gomega.Equal(1))
},
	ginkgo.Entry("pull", opstats.Operation("docker: image pull"), func(ctx context.Context) error {
		return CliPull(ctx, "img")
	}),
	ginkgo.Entry("push", opstats.Operation("docker: image push"), func(ctx context.Context) error {
		return CliPushWithRetries(ctx, "img")
	}),
)

var _ = ginkgo.Describe("docker cli build", func() {
	ginkgo.It("measures the cli build itself", func(ctx ginkgo.SpecContext) {
		cliCtx := cliDaemonContext(newMissingImageDaemonServer())

		collector := opstats.NewCollector()
		err := CliBuild_LiveOutputWithCustomIn(opstats.NewContext(cliCtx, collector), io.NopCloser(strings.NewReader("")), "-")
		gomega.Expect(err).To(gomega.HaveOccurred())

		summary := collector.Summary()
		gomega.Expect(summary).To(gomega.HaveLen(1))
		gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.Operation("docker: image build")))
		gomega.Expect(summary[0].Count).To(gomega.Equal(1))
	})
})

var _ = ginkgo.Describe("docker cli run", func() {
	const createDelay = 300 * time.Millisecond

	ginkgo.It("measures the whole cli run, not just the container start request", func(ctx ginkgo.SpecContext) {
		cliCtx := cliDaemonContext(newDelayedCreateServer(createDelay))

		collector := opstats.NewCollector()
		gomega.Expect(CliRun(opstats.NewContext(cliCtx, collector), "img")).NotTo(gomega.Succeed())

		summary := collector.Summary()
		gomega.Expect(summary).To(gomega.HaveLen(1))
		gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.Operation("docker: container run")))
		gomega.Expect(summary[0].Count).To(gomega.Equal(1))
		gomega.Expect(summary[0].TotalTime).To(gomega.BeNumerically(">=", createDelay), "the timer must cover the container create request too")
	})
})
