package import_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/stapel"
)

var _ = ginkgo.Describe("Import server lifecycle", func() {
	ginkgo.DescribeTable("validates the cached endpoint through Docker inspect", func(ctx ginkgo.SpecContext, status int, body string, healthy, failed bool) {
		inspectCtx := importServerDockerContext(ctx, status, body)
		srv := &RsyncServer{DockerContainerName: "owned-import-server", IPAddress: "172.17.0.25"}
		result, err := srv.Healthy(inspectCtx)
		gomega.Expect(result).To(gomega.Equal(healthy))
		gomega.Expect(err != nil).To(gomega.Equal(failed))
	}, ginkgo.Entry("running", 200, `{"State":{"Running":true},"NetworkSettings":{"Networks":{"bridge":{"IPAddress":"172.17.0.25"}}}}`, true, false),
		ginkgo.Entry("removed", 404, `{"message":"No such container"}`, false, false),
		ginkgo.Entry("stopped", 200, `{"State":{"Running":false},"NetworkSettings":{"Networks":{"bridge":{"IPAddress":"172.17.0.25"}}}}`, false, false),
		ginkgo.Entry("detached", 200, `{"State":{"Running":true},"NetworkSettings":{"Networks":{}}}`, false, false),
		ginkgo.Entry("changed address", 200, `{"State":{"Running":true},"NetworkSettings":{"Networks":{"bridge":{"IPAddress":"172.17.0.26"}}}}`, false, false),
		ginkgo.Entry("daemon error", 500, `{"message":"daemon unavailable"}`, false, true))

	ginkgo.It("tolerates an already removed owned server during cleanup", func(ctx ginkgo.SpecContext) {
		inspectCtx := importServerDockerContext(ctx, http.StatusNotFound, `{"message":"No such container"}`)
		directory := filepath.Join(ginkgo.GinkgoT().TempDir(), "generation")
		gomega.Expect(os.Mkdir(directory, 0o700)).To(gomega.Succeed())
		srv := &RsyncServer{DockerContainerName: "owned-import-server", runtimeDir: directory}
		gomega.Expect(srv.Shutdown(inspectCtx)).To(gomega.Succeed())
		_, err := os.Stat(directory)
		gomega.Expect(errors.Is(err, os.ErrNotExist)).To(gomega.BeTrue())
	})

	ginkgo.It("checks cancellation before inspecting", func(ctx ginkgo.SpecContext) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		srv := &RsyncServer{DockerContainerName: "owned-import-server"}
		_, err := srv.Healthy(canceled)
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))
	})

	ginkgo.DescribeTable("keeps stat/copy status and suppresses password tracing", func(ctx ginkgo.SpecContext, statCode, copyCode, want int) {
		root := ginkgo.GinkgoT().TempDir()
		fake := filepath.Join(root, "rsync")
		script := fmt.Sprintf(`#!/bin/sh
if [ "$1" = -L ]; then printf '%%s\n' '-rw-r--r-- file'; exit %d; fi
printf copied
exit %d
`, statCode, copyCode)
		gomega.Expect(os.WriteFile(fake, []byte(script), 0o700)).To(gomega.Succeed())
		srv := &RsyncServer{IPAddress: "172.17.0.25", Port: "873", AuthUser: "fixture-user", AuthPassword: "fixture-secret"}
		command := srv.GetCopyCommand(ctx, &config.Import{ArtifactExport: &config.ArtifactExport{ExportBase: &config.ExportBase{Add: "/source", To: filepath.Join(root, "target")}}})
		gomega.Expect(command).To(gomega.ContainSubstring("fixture-secret"))
		command = strings.NewReplacer(stapel.RsyncBinPath(), fake, stapel.HeadBinPath(), "head", stapel.MkdirBinPath(), "mkdir").Replace(command)
		process := exec.CommandContext(ctx, "bash", "-ec", `script=$(</dev/stdin); eval "$script" < /dev/null`)
		process.Stdin = strings.NewReader("set -x && " + command)
		output, err := process.CombinedOutput()
		if want == 0 {
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), string(output))
			gomega.Expect(string(output)).To(gomega.ContainSubstring("copied"))
		} else {
			var exit *exec.ExitError
			gomega.Expect(errors.As(err, &exit)).To(gomega.BeTrue())
			gomega.Expect(exit.ExitCode()).To(gomega.Equal(want))
		}
		gomega.Expect(string(output)).To(gomega.ContainSubstring("set +x"))
		gomega.Expect(string(output)).NotTo(gomega.ContainSubstring("fixture-secret"))
	}, ginkgo.Entry("successful copy", 0, 0, 0), ginkgo.Entry("stat transport failure", 10, 0, 10),
		ginkgo.Entry("copy transport failure", 0, 10, 10), ginkgo.Entry("protocol stream failure", 0, 12, 12), ginkgo.Entry("partial transfer", 0, 23, 23))
})
