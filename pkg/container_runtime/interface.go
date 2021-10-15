package container_runtime

import (
	"context"
	"io"

	"github.com/werf/werf/pkg/image"
)

type BuildDockerfileOptions struct {
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

type ContainerRuntime interface {
	Tag(ctx context.Context, ref, newRef string) error
	Push(ctx context.Context, ref string) error
	Pull(ctx context.Context, ref string) error
	Rmi(ctx context.Context, ref string) error

	GetImageInfo(ctx context.Context, ref string) (*image.Info, error)
	BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOptions) (string, error)
	// StapelBuild(opts StapelBuildOptions) string

	String() string

	// Legacy
	RefreshImageObject(ctx context.Context, img LegacyImageInterface) error
	PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error
	RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img LegacyImageInterface) error
}
