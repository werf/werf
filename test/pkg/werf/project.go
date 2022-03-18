package werf

import (
	"encoding/json"
	"os"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build"
	iutils "github.com/werf/werf/test/pkg/utils"
)

func NewProject(werfBinPath, gitRepoPath string) *Project {
	return &Project{
		WerfBinPath: werfBinPath,
		GitRepoPath: gitRepoPath,
	}
}

type Project struct {
	GitRepoPath string
	WerfBinPath string
}

type CommonOptions struct {
	ShouldFail bool
	ExtraArgs  []string
}

type BuildOptions struct {
	CommonOptions
}

type BuildWithReportOptions struct {
	CommonOptions
}

type KubeRunOptions struct {
	CommonOptions
	Command string
	Image   string
}

type KubeCtlOptions struct {
	CommonOptions
}

type runCommandOptions struct {
	ShouldFail bool
	Args       []string
}

func (p *Project) Build(opts *BuildOptions) (combinedOut string) {
	if opts == nil {
		opts = &BuildOptions{}
	}

	args := append([]string{"build"}, opts.ExtraArgs...)
	outb, err := iutils.RunCommand(p.GitRepoPath, p.WerfBinPath, args...)
	if opts.ShouldFail {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).NotTo(HaveOccurred())
	}

	return string(outb)
}

func (p *Project) BuildWithReport(buildReportPath string, opts *BuildWithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &BuildWithReportOptions{}
	}

	args := append([]string{"build", "--report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) KubeRun(opts *KubeRunOptions) string {
	if opts == nil {
		opts = &KubeRunOptions{}
	}

	args := append([]string{"kube-run"}, opts.ExtraArgs...)

	if opts.Image != "" {
		args = append(args, opts.Image)
	}

	if opts.Command != "" {
		args = append(args, "--", opts.Command)
	}

	return p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})
}

func (p *Project) KubeCtl(opts *KubeCtlOptions) string {
	if opts == nil {
		opts = &KubeCtlOptions{}
	}

	args := append([]string{"kubectl"}, opts.ExtraArgs...)

	return p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})
}

func (p *Project) runCommand(opts runCommandOptions) string {
	outb, err := iutils.RunCommand(p.GitRepoPath, p.WerfBinPath, opts.Args...)
	if opts.ShouldFail {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).NotTo(HaveOccurred())
	}

	return string(outb)
}
