package container_backend

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/stapel"
)

var _ = Describe("Legacy stage command transport", func() {
	var container *LegacyStageImageContainer

	BeforeEach(func() {
		container = newLegacyStageImageContainer(&LegacyStageImage{
			targetPlatform: "linux/amd64",
			fromImage: &LegacyStageImage{
				legacyBaseImage: &legacyBaseImage{name: "base"},
			},
		})
	})

	It("keeps large scripts intact and outside argv", func(ctx SpecContext) {
		commands := []string{strings.Repeat(": import-padding\n", 200_000) + ":", "printf complete"}
		container.AddRunCommands(commands...)

		Expect(container.prepareRunCommand(ctx)).To(Equal(strings.Join(append(container.prepareBuildTimeEnvExports(ctx), commands...), " && ")))
		args := container.prepareRunCommandArgs()
		Expect(args[:3]).To(Equal([]string{"-i", "base", "-ec"}))
		Expect(args[3]).NotTo(ContainSubstring(stapel.CONTAINER_MOUNT_ROOT), "the script loader must not depend on a binary of the pinned stapel image")
		for _, arg := range args {
			Expect(len(arg)).To(BeNumerically("<", 128*1024))
			Expect(arg).NotTo(ContainSubstring("import-padding"))
		}
	})

	It("loads the script with no executable available", func(ctx SpecContext) {
		args := container.prepareRunCommandArgs()
		cmd := exec.CommandContext(ctx, "bash", args[2], args[3])
		cmd.Env = append(os.Environ(), "PATH=")
		cmd.Stdin = strings.NewReader("printf complete")

		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(string(output)).To(Equal("complete"))
	})

	DescribeTable("propagates script reader failures before evaluation", func(ctx SpecContext, reader string, exitCode int) {
		args := container.prepareRunCommandArgs()
		loader := strings.ReplaceAll(args[3], "</dev/stdin", reader)
		cmd := exec.CommandContext(ctx, "bash", args[2], loader)
		cmd.Stdin = strings.NewReader("cat > /dev/null; printf complete")
		output, err := cmd.Output()
		if exitCode == 0 {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal("complete"))
			return
		}
		Expect(err).To(HaveOccurred())
		var exitErr *exec.ExitError
		Expect(errors.As(err, &exitErr)).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(exitCode))
		Expect(output).To(BeEmpty())
	},
		Entry("successful reader", "cat", 0),
		Entry("failed reader", "false", 1),
		Entry("failed reader with partial script", "printf 'printf unexpected'; false", 1),
	)

	It("prints a runnable debug command preserving stdin and argument quoting", func(ctx SpecContext) {
		script := "printf '%s\\n' '$HOME' \"a'b\"; $(not-a-host-command)\n"
		container.AddRunCommands(script)
		args := []string{"-i", "--env=VALUE=spaces 'quotes' $HOME", "base", "-ec", "reader; eval"}
		cmd := exec.CommandContext(ctx, "bash", "-c", "docker() { printf '%s\\n' \"$@\"; cat; }\n"+container.prepareDebugRunCommand(ctx, args))
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		expectedScript := strings.Join(append(container.prepareBuildTimeEnvExports(ctx), script), " && ")
		Expect(string(output)).To(Equal("run\n" + strings.Join(args, "\n") + "\n" + expectedScript))
	})
})
