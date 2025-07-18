package container_backend

import (
	"bytes"
	"context"
	"io"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
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
	DockerfileCtxRelPath string
	Target               string
	BuildArgs            []string // {"key1=value1", "key2=value2", ... }
	AddHost              []string
	Network              string
	SSH                  string
	Labels               []string
	Tags                 []string
	Secrets              []string
	Quiet                bool
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

//go:generate mockgen -source interface.go -package mock -destination ../../test/mock/container_backend.go

type ContainerBackend interface {
	Info(ctx context.Context) (info.Info, error)

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

	// PruneImages removes all dangling images
	PruneImages(ctx context.Context, options prune.Options) (prune.Report, error)
	// PruneVolumes removes all anonymous volumes not used by at least one container
	PruneVolumes(ctx context.Context, options prune.Options) (prune.Report, error)

	// DumpImage streams image using bytes reader
	DumpImage(ctx context.Context, ref string) (*bytes.Reader, error)

	// GenerateSBOM scans and generates SBOM from source image into another destination image
	GenerateSBOM(ctx context.Context, scanOpts scanner.ScanOptions, dstImgLabels []string) (string, error)

	String() string

	// TODO: Util method for cleanup, which possibly should be avoided in the future
	RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error

	// Legacy
	RefreshImageObject(ctx context.Context, img LegacyImageInterface) error
	PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error
	RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img LegacyImageInterface) error
	TagImageByName(ctx context.Context, img LegacyImageInterface) error
	// Mutation
	SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error)
	LoadImageFromStream(ctx context.Context, input io.Reader) (string, error)
}

func PullImageFromRegistry(ctx context.Context, containerBackend ContainerBackend, img LegacyImageInterface) error {
	return logboek.Context(ctx).Info().
		LogProcess("Pulling image %s", img.Name()).
		DoError(func() error {
			return SanitizeError(
				containerBackend.PullImageFromRegistry(ctx, img),
			)
		})
}
