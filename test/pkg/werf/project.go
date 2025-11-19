package werf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/gomega"

	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/werf/v2/pkg/build"
	iutils "github.com/werf/werf/v2/test/pkg/utils"
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

	CancelOnOutput        string
	CancelOnOutputTimeout time.Duration
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

type BundlePublishOptions struct {
	CommonOptions
}

type BundlePublishWithReportOptions struct {
	CommonOptions
}

type BundleApplyOptions struct {
	CommonOptions
}

type BundleApplyWithReportOptions struct {
	CommonOptions
}

type ExportOptions struct {
	CommonOptions
}

type SbomGetOptions struct {
	CommonOptions
}

type VerifyOptions struct {
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
	Envs       []string

	CancelOnOutput        string
	CancelOnOutputTimeout time.Duration
}

func (p *Project) Build(ctx context.Context, opts *BuildOptions) (combinedOut string) {
	if opts == nil {
		opts = &BuildOptions{}
	}

	args := append([]string{"build"}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) BuildWithReport(ctx context.Context, buildReportPath string, opts *BuildWithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &BuildWithReportOptions{}
	}

	args := append([]string{"build", "--save-build-report", "--build-report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) Converge(ctx context.Context, opts *ConvergeOptions) (combinedOut string) {
	if opts == nil {
		opts = &ConvergeOptions{}
	}

	args := append([]string{"converge"}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) ConvergeWithReport(ctx context.Context, deployReportPath string, opts *ConvergeWithReportOptions) (combinedOut string, report release.DeployReport) {
	if opts == nil {
		opts = &ConvergeWithReportOptions{}
	}

	args := append([]string{"converge", "--save-deploy-report", "--deploy-report-path", deployReportPath}, opts.ExtraArgs...)
	out := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	deployReportRaw, err := os.ReadFile(deployReportPath)
	Expect(err).NotTo(HaveOccurred())

	var deployReport release.DeployReport
	Expect(json.Unmarshal(deployReportRaw, &deployReport)).To(Succeed())

	return out, deployReport
}

func (p *Project) BundlePublish(ctx context.Context, opts *BundlePublishOptions) (combinedOut string) {
	if opts == nil {
		opts = &BundlePublishOptions{}
	}

	args := append([]string{"bundle", "publish"}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) BundlePublishWithReport(ctx context.Context, buildReportPath string, opts *BundlePublishWithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &BundlePublishWithReportOptions{}
	}

	args := append([]string{"bundle", "publish", "--save-build-report", "--build-report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) BundleApply(ctx context.Context, releaseName, namespace string, opts *BundleApplyOptions) (combinedOut string) {
	if opts == nil {
		opts = &BundleApplyOptions{}
	}

	args := append([]string{"bundle", "apply", "--release", releaseName, "--namespace", namespace}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) BundleApplyWithReport(ctx context.Context, releaseName, namespace, deployReportPath string, opts *BundleApplyWithReportOptions) (string, release.DeployReport) {
	if opts == nil {
		opts = &BundleApplyWithReportOptions{}
	}

	args := append([]string{"bundle", "apply", "--release", releaseName, "--namespace", namespace, "--save-deploy-report", "--deploy-report-path", deployReportPath}, opts.ExtraArgs...)
	out := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	deployReportRaw, err := os.ReadFile(deployReportPath)
	Expect(err).NotTo(HaveOccurred())

	var deployReport release.DeployReport
	Expect(json.Unmarshal(deployReportRaw, &deployReport)).To(Succeed())

	return out, deployReport
}

func (p *Project) KubeRun(ctx context.Context, opts *KubeRunOptions) string {
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

	return p.runCommand(ctx, runCommandOptions{
		Args:                  args,
		ShouldFail:            opts.ShouldFail,
		CancelOnOutput:        opts.CancelOnOutput,
		CancelOnOutputTimeout: opts.CancelOnOutputTimeout,
	})
}

func (p *Project) KubeCtl(ctx context.Context, opts *KubeCtlOptions) string {
	if opts == nil {
		opts = &KubeCtlOptions{}
	}

	args := append([]string{"kubectl"}, opts.ExtraArgs...)

	return p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})
}

func (p *Project) Namespace(ctx context.Context) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.namespace == "" {
		p.namespace = strings.TrimSpace(p.runCommand(ctx, runCommandOptions{Args: []string{"helm", "get-namespace"}}))
	}

	return p.namespace
}

func (p *Project) Release(ctx context.Context) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.release == "" {
		p.release = strings.TrimSpace(p.runCommand(ctx, runCommandOptions{Args: []string{"helm", "get-release"}}))
	}

	return p.release
}

func (p *Project) CreateNamespace(ctx context.Context) {
	if getNsOut := p.KubeCtl(ctx, &KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"get", "namespace", "--ignore-not-found", p.Namespace(ctx),
			},
		},
	}); getNsOut != "" {
		return
	}

	p.KubeCtl(ctx, &KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"create", "namespace", p.Namespace(ctx),
			},
		},
	})
}

func (p *Project) CreateRegistryPullSecretFromDockerConfig(ctx context.Context) {
	user, err := user.Current()
	Expect(err).NotTo(HaveOccurred())

	p.KubeCtl(ctx, &KubeCtlOptions{
		CommonOptions: CommonOptions{
			ExtraArgs: []string{
				"create", "secret", "docker-registry", "registry", "-n", p.Namespace(ctx),
				"--from-file", fmt.Sprintf(".dockerconfigjson=%s", filepath.Join(user.HomeDir, ".docker", "config.json")),
			},
		},
	})
}

func (p *Project) runCommand(ctx context.Context, opts runCommandOptions) string {
	outb, _ := iutils.RunCommandWithOptions(ctx, p.GitRepoPath, p.WerfBinPath, opts.Args, iutils.RunCommandOptions{
		ShouldSucceed:         !opts.ShouldFail,
		ExtraEnv:              opts.Envs,
		CancelOnOutput:        opts.CancelOnOutput,
		CancelOnOutputTimeout: opts.CancelOnOutputTimeout,
	})
	return string(outb)
}

func (p *Project) Compose(ctx context.Context, opts *BuildOptions) (combinedOut string) {
	if opts == nil {
		opts = &BuildOptions{}
	}

	args := append([]string{"compose"}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) Export(ctx context.Context, opts *ExportOptions) (combinedOut string) {
	if opts == nil {
		opts = &ExportOptions{}
	}
	args := append([]string{"export"}, opts.ExtraArgs...)
	outb := p.runCommand(ctx, runCommandOptions{Args: args, ShouldFail: opts.ShouldFail})

	return string(outb)
}

func (p *Project) SbomGet(ctx context.Context, opts *SbomGetOptions) (combinedOut string) {
	if opts == nil {
		opts = &SbomGetOptions{}
	}
	args := append([]string{"sbom", "get"}, opts.ExtraArgs...)

	outb := p.runCommand(ctx, runCommandOptions{
		Args:       args,
		ShouldFail: opts.ShouldFail,
	})

	return string(outb)
}

func (p *Project) Verify(ctx context.Context, opts *VerifyOptions) (combinedOut string) {
	if opts == nil {
		opts = &VerifyOptions{}
	}
	args := append([]string{"verify"}, opts.ExtraArgs...)

	outb := p.runCommand(ctx, runCommandOptions{
		Args:       args,
		ShouldFail: opts.ShouldFail,
	})

	return string(outb)
}
