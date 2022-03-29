package container_runtime

import (
	"context"
	"io"

	"github.com/werf/werf/pkg/image"
)

type CommonOpts struct{}

type TagOpts struct {
	CommonOpts
}

type PushOpts struct {
	CommonOpts
}

type PullOpts struct {
	CommonOpts
}

type RmiOpts struct {
	CommonOpts
}

type GetImageInfoOpts struct {
	CommonOpts
}

type BuildDockerfileOpts struct {
	CommonOpts

	ContextTar           io.ReadCloser
	DockerfileCtxRelPath string // TODO: remove this and instead write the []byte dockerfile to /Dockerfile in the ContextTar inDockerServerRuntime.BuildDockerfile().
	Target               string
	BuildArgs            []string // {"key1=value1", "key2=value2", ... }
	AddHost              []string
	Network              string
	SSH                  string
	Labels               []string
	Tags                 []string
}

// type StapelBuildOptions struct {
//	ServiceRunCommands []string
//	RunCommands []string
//	Volumes []string
//	VolumesFrom []string
//	Exposes []string
//	Envs map[string]string
//	Labels map[string]string
// }

type BuildStapelStageOpts struct {
	BuildVolumes            []string
	Labels                  []string
	UserCommands            []string
	PrepareContainerActions []PrepareContainerAction
}

type PrepareContainerAction interface {
	PrepareContainer(containerRoot string) error
}

type PrepareContainerActionWith func(containerRoot string) error

func (f PrepareContainerActionWith) PrepareContainer(containerRoot string) error {
	return f(containerRoot)
}

type ContainerRuntime interface {
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	Pull(ctx context.Context, ref string, opts PullOpts) error
	Rmi(ctx context.Context, ref string, opts RmiOpts) error

	GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error)
	BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error)
	BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOpts) (string, error)

	HasStapelBuildSupport() bool
	String() string

	// Legacy
	RefreshImageObject(ctx context.Context, img LegacyImageInterface) error
	PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error
	RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img LegacyImageInterface) error
}
