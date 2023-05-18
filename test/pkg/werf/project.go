package werf

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/release"

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

	namespace string
	release   string
	mu        sync.Mutex
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

type ConvergeOptions struct {
	CommonOptions
}

type ConvergeWithReportOptions struct {
	CommonOptions
}

type KubeRunOptions struct {
	CommonOptions
	Command []string
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
	outb := p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) BuildWithReport(buildReportPath string, opts *BuildWithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &BuildWithReportOptions{}
	}

	args := append([]string{"build", "--save-build-report", "--build-report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) Converge(opts *ConvergeOptions) (combinedOut string) {
	if opts == nil {
		opts = &ConvergeOptions{}
	}

	args := append([]string{"converge"}, opts.ExtraArgs...)
	outb := p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) ConvergeWithReport(deployReportPath string, opts *ConvergeWithReportOptions) (combinedOut string, report release.DeployReport) {
	if opts == nil {
		opts = &ConvergeWithReportOptions{}
	}

	args := append([]string{"converge", "--save-deploy-report", "--deploy-report-path", deployReportPath}, opts.ExtraArgs...)
	out := p.runCommand(runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	deployReportRaw, err := os.ReadFile(deployReportPath)
	Expect(err).NotTo(HaveOccurred())

	var deployReport release.DeployReport
	Expect(json.Unmarshal(deployReportRaw, &deployReport)).To(Succeed())

	return out, deployReport
}

func (p *Project) KubeRun(opts *KubeRunOptions) string {
	if opts == nil {
		opts = &KubeRunOptions{}
	}

	args := append([]string{"kube-run"}, opts.ExtraArgs...)

	if opts.Image != "" {
		args = append(args, opts.Image)
	}

	if len(opts.Command) > 0 {
		args = append(args, "--")
		args = append(args, opts.Command...)
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

func (p *Project) Namespace() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.namespace == "" {
		p.namespace = strings.TrimSpace(p.runCommand(runCommandOptions{Args: []string{"helm", "get-namespace"}}))
	}

	return p.namespace
}

func (p *Project) Release() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.release == "" {
		p.release = strings.TrimSpace(p.runCommand(runCommandOptions{Args: []string{"helm", "get-release"}}))
	}

	return p.release
}

func (p *Project) CreateNamespace() {
	if getNsOut := p.KubeCtl(&KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"get", "namespace", "--ignore-not-found", p.Namespace(),
			},
		},
	}); getNsOut != "" {
		return
	}

	p.KubeCtl(&KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"create", "namespace", p.Namespace(),
			},
		},
	})
}

func (p *Project) CreateRegistryPullSecretFromDockerConfig() {
	user, err := user.Current()
	Expect(err).NotTo(HaveOccurred())

	p.KubeCtl(&KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"create", "secret", "docker-registry", "registry", "-n", p.Namespace(),
				"--from-file", fmt.Sprintf(".dockerconfigjson=%s", filepath.Join(user.HomeDir, ".docker", "config.json")),
			},
		},
	})
}

func (p *Project) runCommand(opts runCommandOptions) string {
	outb, err := iutils.RunCommand(p.GitRepoPath, p.WerfBinPath, iutils.WerfBinArgs(opts.Args...)...)
	if opts.ShouldFail {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).NotTo(HaveOccurred())
	}

	return string(outb)
}
