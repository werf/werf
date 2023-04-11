package container_backend

import (
	"context"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type CommonOpts struct {
	TargetPlatform string
}

type (
	TagOpts                           CommonOpts
	PushOpts                          CommonOpts
	PullOpts                          CommonOpts
	GetImageInfoOpts                  CommonOpts
	CalculateDependencyImportChecksum CommonOpts
)

type RmOpts struct {
	CommonOpts
	Force bool
}

type RmiOpts struct {
	CommonOpts
	Force bool
}

type BuildDockerfileOpts struct {
	CommonOpts

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

type ImagesOptions struct {
	CommonOpts
	Filters []util.Pair[string, string]
}

type ContainersOptions struct {
	CommonOpts
	Filters []image.ContainerFilter
}

type PostManifestOpts struct {
	CommonOpts
	Labels    []string
	Manifests []*image.Info
}

type ContainerBackend interface {
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	Pull(ctx context.Context, ref string, opts PullOpts) error
	Rmi(ctx context.Context, ref string, opts RmiOpts) error
	Rm(ctx context.Context, name string, opts RmOpts) error
	PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error

	GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error)
	BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error)
	BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error)
	BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error)
	CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (string, error)

	HasStapelBuildSupport() bool
	GetDefaultPlatform() string
	GetRuntimePlatform() string

	Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error)
	Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error)

	ClaimTargetPlatforms(ctx context.Context, targetPlatforms []string)

	String() string

	// TODO: Util method for cleanup, which possibly should be avoided in the future
	RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error

	// Legacy
	ShouldCleanupDockerfileImage() bool
	RefreshImageObject(ctx context.Context, img LegacyImageInterface) error
	PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error
	RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img LegacyImageInterface) error
	TagImageByName(ctx context.Context, img LegacyImageInterface) error
}
