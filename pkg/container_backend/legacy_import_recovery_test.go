package container_backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Import stage recovery", func() {
	ginkgo.DescribeTable("classifies typed transport failures", func(code int, expected bool) {
		err := fmt.Errorf("container run failed: %w", cli.StatusError{StatusCode: code})
		gomega.Expect(isImportTransportError(err)).To(gomega.Equal(expected))
	}, ginkgo.Entry("socket", 10, true), ginkgo.Entry("stream", 12, true),
		ginkgo.Entry("timeout", 30, true), ginkgo.Entry("connect timeout", 35, true),
		ginkgo.Entry("file IO", 11, false), ginkgo.Entry("partial transfer", 23, false),
		ginkgo.Entry("Docker startup", 125, false))

	ginkgo.It("does not classify error text as an exit status", func() {
		gomega.Expect(isImportTransportError(errors.New("container run failed: exit code 10"))).To(gomega.BeFalse())
	})

	ginkgo.It("refreshes commands and discards partial contents before one retry", func(ctx ginkgo.SpecContext) {
		destination := filepath.Join(ginkgo.GinkgoT().TempDir(), "destination")
		generation, attempts := 0, 0
		err := (&LegacyStageImageContainer{}).runImportAttempts(ctx, func(context.Context) error { generation++; return nil }, func(context.Context) error {
			attempts++
			if attempts == 1 {
				gomega.Expect(os.WriteFile(destination, []byte("partial"), 0o600)).To(gomega.Succeed())
				return cli.StatusError{StatusCode: 10}
			}
			gomega.Expect(generation).To(gomega.Equal(2))
			_, err := os.Stat(destination)
			gomega.Expect(errors.Is(err, os.ErrNotExist)).To(gomega.BeTrue())
			return os.WriteFile(destination, []byte("complete"), 0o600)
		}, func(context.Context) error { return os.Remove(destination) })
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		data, err := os.ReadFile(destination)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(string(data)).To(gomega.Equal("complete"))
		gomega.Expect(attempts).To(gomega.Equal(2))
	})

	ginkgo.DescribeTable("preserves the original error and bounds recovery", func(ctx ginkgo.SpecContext, failure string, expectedRuns, expectedPrepares int) {
		callCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		runs, prepares := 0, 0
		original := fmt.Errorf("container run failed: %w", cli.StatusError{StatusCode: 10})
		if failure == "data" {
			original = fmt.Errorf("original: %w", cli.StatusError{StatusCode: 23})
		}
		container := &LegacyStageImageContainer{}
		err := container.runImportAttempts(callCtx, func(context.Context) error {
			prepares++
			if prepares == 2 && failure == "refresh" {
				return errors.New("refresh failed")
			}
			return nil
		}, func(context.Context) error {
			container.importDestinationDiscarded = false
			runs++
			if failure == "cancel" {
				cancel()
			}
			return original
		}, func(context.Context) error {
			if failure == "remove" {
				return errors.New("remove conflict")
			}
			return nil
		})
		gomega.Expect(errors.Is(err, original)).To(gomega.BeTrue())
		gomega.Expect(runs).To(gomega.Equal(expectedRuns))
		gomega.Expect(prepares).To(gomega.Equal(expectedPrepares))
		gomega.Expect(container.importDestinationDiscarded).To(gomega.Equal(failure == "refresh"))
		if failure != "data" {
			gomega.Expect(err.Error()).To(gomega.HavePrefix("container run failed"))
		}
	}, ginkgo.Entry("repeated transport failure", "repeat", 2, 2),
		ginkgo.Entry("data failure", "data", 1, 1), ginkgo.Entry("canceled worker", "cancel", 1, 1),
		ginkgo.Entry("replacement failure", "refresh", 1, 2), ginkgo.Entry("removal conflict", "remove", 1, 1))

	ginkgo.It("does no preparation after cancellation", func(ctx ginkgo.SpecContext) {
		callCtx, cancel := context.WithCancel(ctx)
		cancel()
		called := false
		operation := func(context.Context) error { called = true; return nil }
		gomega.Expect((&LegacyStageImageContainer{}).runImportAttempts(callCtx, operation, operation, operation)).To(gomega.MatchError(context.Canceled))
		gomega.Expect(called).To(gomega.BeFalse())
	})

	ginkgo.It("replaces the import block and excludes it from diagnostics", func(ctx ginkgo.SpecContext) {
		c := newLegacyStageImageContainer(&LegacyStageImage{})
		c.importCommands = []string{"(set +x; RSYNC_PASSWORD='fixture-secret' printf copied)"}
		c.AddServiceRunCommands(": service")
		c.AddRunCommands(": user")
		gomega.Expect(c.prepareAllRunCommands()).To(gomega.ContainElement(c.importCommands[0]))
		gomega.Expect(c.diagnosticRunCommands()).To(gomega.Equal([]string{": service", ": '[import command redacted]'", ": user"}))
		c.importCommands = []string{"(set +x; RSYNC_PASSWORD='new-secret' printf copied)"}
		gomega.Expect(strings.Join(c.prepareAllRunCommands(), "\n")).NotTo(gomega.ContainSubstring("fixture-secret"))
		gomega.Expect(c.prepareAllRunCommands()).To(gomega.HaveLen(3))
	})

	ginkgo.DescribeTable("keeps inherited shell tracing away from import credentials", func(ctx ginkgo.SpecContext, options string) {
		c := newLegacyStageImageContainer(&LegacyStageImage{fromImage: &LegacyStageImage{legacyBaseImage: &legacyBaseImage{name: "base"}}})
		c.importCommands = []string{"(set +x; RSYNC_PASSWORD='fixture-secret' printf copied)"}
		args := c.prepareRunCommandArgs()
		process := exec.CommandContext(ctx, "bash", args[2], args[3])
		process.Env = append(os.Environ(), "SHELLOPTS="+options)
		process.Stdin = strings.NewReader(c.prepareRunCommand(ctx))
		output, err := process.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), string(output))
		gomega.Expect(string(output)).To(gomega.ContainSubstring("copied"))
		gomega.Expect(string(output)).NotTo(gomega.ContainSubstring("fixture-secret"))
	}, ginkgo.Entry("xtrace", "xtrace"), ginkgo.Entry("verbose", "verbose"), ginkgo.Entry("both", "xtrace:verbose"))

	ginkgo.It("keeps the destination discarded when retry preparation fails before Docker run", func() {
		ctx, _ := dockerDaemonContext(daemonHandler(http.StatusNotFound, `{"message":"No such image: pinned-base"}`))
		backend := &DockerServerBackend{}
		base := NewLegacyStageImage(&LegacyStageImage{}, "base", backend, "linux/amd64")
		stage := NewLegacyStageImage(base, "destination", backend, "linux/amd64")
		container := stage.container
		container.baseImageID = "pinned-base"
		runs := 0
		err := container.runImportAttempts(ctx,
			func(context.Context) error { return nil },
			func(ctx context.Context) error {
				runs++
				if runs == 1 {
					return fmt.Errorf("container run failed: %w", cli.StatusError{StatusCode: 12})
				}
				return container.run(ctx)
			},
			func(context.Context) error { return nil })
		gomega.Expect(runs).To(gomega.Equal(2))
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("image pinned-base is not available locally")))
		gomega.Expect(container.importDestinationDiscarded).To(gomega.BeTrue())
	})

	ginkgo.DescribeTable("does not replay mixed command stages", func(ctx ginkgo.SpecContext, service bool) {
		container := &LegacyStageImageContainer{}
		if service {
			container.AddServiceRunCommands("exit 10")
		} else {
			container.AddRunCommands("exit 10")
		}
		prepares, runs, discards := 0, 0, 0
		original := cli.StatusError{StatusCode: 10}
		err := container.runImportAttempts(ctx,
			func(context.Context) error { prepares++; return nil },
			func(context.Context) error { runs++; return original },
			func(context.Context) error { discards++; return nil })
		gomega.Expect(err).To(gomega.Equal(original))
		gomega.Expect(prepares).To(gomega.Equal(1))
		gomega.Expect(runs).To(gomega.Equal(1))
		gomega.Expect(discards).To(gomega.BeZero())
	}, ginkgo.Entry("user commands", false), ginkgo.Entry("service commands", true))
})
