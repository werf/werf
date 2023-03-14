package container_backend

import (
	"context"

	"github.com/werf/werf/pkg/image"
)

type CommonOpts struct {
	TargetPlatform string
}

type (
	TagOpts                           CommonOpts
	PushOpts                          CommonOpts
	PullOpts                          CommonOpts
	RmiOpts                           CommonOpts
	GetImageInfoOpts                  CommonOpts
	CalculateDependencyImportChecksum CommonOpts
)

type BuildDockerfileOpts struct {
	CommonOpts

	TargetPlatform       string
	BuildContextArchive  BuildContextArchiver
	DockerfileCtxRelPath string // TODO: remove this and instead write the []byte dockerfile to /Dockerfile in the ContextTar inDockerServerBackend.BuildDockerfile().
	Target               string
	BuildArgs            []string // {"key1=value1", "key2=value2", ... }
	AddHost              []string
	Network              string
	SSH                  string
	Labels               []string
	Tags                 []string
}

type BuildDockerfileStageOptions struct {
	CommonOpts

	BuildContextArchive BuildContextArchiver
}

type BuildOptions struct {
	TargetPlatform        string
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type ContainerBackend interface {
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	Pull(ctx context.Context, ref string, opts PullOpts) error
	Rmi(ctx context.Context, ref string, opts RmiOpts) error

	GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error)
	BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error)
	BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error)
	BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error)
	CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (string, error)

	HasStapelBuildSupport() bool
	GetDefaultPlatform() string
	GetRuntimePlatform() string

	String() string

	// TODO: Util method for cleanup, which possibly should be avoided in the future
	RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error

	// Legacy
	ShouldCleanupDockerfileImage() bool
	RefreshImageObject(ctx context.Context, img LegacyImageInterface) error
	PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error
	RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img LegacyImageInterface) error
}
