package werf

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/onsi/gomega"

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

func (p *Project) Build(ctx context.Context, opts *BuildOptions) (combinedOut string) {
	if opts == nil {
		opts = &BuildOptions{}
	}

	args := append([]string{"build"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) Converge(ctx context.Context, opts *ConvergeOptions) (combinedOut string) {
	if opts == nil {
		opts = &ConvergeOptions{}
	}

	args := append([]string{"converge"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) BundlePublish(ctx context.Context, opts *BundlePublishOptions) (combinedOut string) {
	if opts == nil {
		opts = &BundlePublishOptions{}
	}

	args := append([]string{"bundle", "publish"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) BundleApply(ctx context.Context, releaseName, namespace string, opts *BundleApplyOptions) (combinedOut string) {
	if opts == nil {
		opts = &BundleApplyOptions{}
	}

	args := append([]string{"bundle", "apply", "--release", releaseName, "--namespace", namespace}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
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

	return p.RunCommand(ctx, args, opts.CommonOptions)
}

func (p *Project) KubeCtl(ctx context.Context, opts *KubeCtlOptions) string {
	if opts == nil {
		opts = &KubeCtlOptions{}
	}

	args := append([]string{"kubectl"}, opts.ExtraArgs...)

	return p.RunCommand(ctx, args, opts.CommonOptions)
}

func (p *Project) Namespace(ctx context.Context) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.namespace == "" {
		p.namespace = strings.TrimSpace(p.RunCommand(ctx, []string{"helm", "get-namespace"}, CommonOptions{}))
	}

	return p.namespace
}

func (p *Project) Release(ctx context.Context) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.release == "" {
		p.release = strings.TrimSpace(p.RunCommand(ctx, []string{"helm", "get-release"}, CommonOptions{}))
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

func (p *Project) RunCommand(
	ctx context.Context,
	args []string,
	opts CommonOptions,
) string {
	outb, _ := iutils.RunCommandWithOptions(
		ctx,
		p.GitRepoPath,
		p.WerfBinPath,
		args,
		iutils.RunCommandOptions{
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
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) Export(ctx context.Context, opts *ExportOptions) (combinedOut string) {
	if opts == nil {
		opts = &ExportOptions{}
	}
	args := append([]string{"export"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) CiEnv(ctx context.Context, opts *CiEnvOptions) (combinedOut string) {
	if opts == nil {
		opts = &CiEnvOptions{}
	}
	args := append([]string{"ci-env"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) HostCleanup(ctx context.Context, opts *HostCleanupOptions) (combinedOut string) {
	if opts == nil {
		opts = &HostCleanupOptions{}
	}
	args := append([]string{"host", "cleanup"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) StagesCopy(ctx context.Context, opts *StagesCopyOptions) (combinedOut string) {
	if opts == nil {
		opts = &StagesCopyOptions{}
	}

	args := append([]string{"stages", "copy"}, opts.ExtraArgs...)
	outb := p.RunCommand(ctx, args, opts.CommonOptions)

	return string(outb)
}

func (p *Project) SbomGet(ctx context.Context, opts *SbomGetOptions) (combinedOut string) {
	if opts == nil {
		opts = &SbomGetOptions{}
	}
	args := append([]string{"sbom", "get"}, opts.ExtraArgs...)

	outb := p.RunCommand(ctx, args, CommonOptions{
		ShouldFail: opts.ShouldFail,
	})

	return string(outb)
}

func (p *Project) Verify(ctx context.Context, opts *VerifyOptions) (combinedOut string) {
	if opts == nil {
		opts = &VerifyOptions{}
	}
	args := append([]string{"verify"}, opts.ExtraArgs...)

	outb := p.RunCommand(ctx, args, CommonOptions{
		ShouldFail: opts.ShouldFail,
	})

	return string(outb)
}
