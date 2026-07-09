package container_backend

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd/platforms"

	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
)

var _ ContainerBackend = (*BuildkitBackend)(nil)

func AsBuildkitBackend(backend ContainerBackend) (*BuildkitBackend, bool) {
	switch b := backend.(type) {
	case *BuildkitBackend:
		return b, true
	case *PerfCheckContainerBackend:
		return AsBuildkitBackend(b.ContainerBackend)
	default:
		return nil, false
	}
}

type BuildkitBackend struct {
	host string
	BuildkitBackendOptions

	mu   sync.Mutex
	repo string
}

type BuildkitBackendOptions struct {
	DefaultPlatform string
	DockerConfigDir string
}

func NewBuildkitBackend(host string, opts BuildkitBackendOptions) *BuildkitBackend {
	return &BuildkitBackend{host: host, BuildkitBackendOptions: opts}
}

func (backend *BuildkitBackend) SetStagesStorageRepo(repo string) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	backend.repo = repo
}

func (backend *BuildkitBackend) getStagesStorageRepo() (string, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if backend.repo == "" {
		return "", fmt.Errorf("stages storage repo is not set: --repo is required when using buildkit backend")
	}
	return backend.repo, nil
}

func (backend *BuildkitBackend) Info(ctx context.Context) (info.Info, error) {
	return info.Info{}, nil
}

func (backend *BuildkitBackend) HasStapelBuildSupport() bool {
	return true
}

func (backend *BuildkitBackend) GetDefaultPlatform() string {
	if backend.DefaultPlatform != "" {
		return backend.DefaultPlatform
	}
	return platforms.Format(platforms.DefaultSpec())
}

func (backend *BuildkitBackend) GetRuntimePlatform() string {
	return platforms.Format(platforms.DefaultSpec())
}

func (backend *BuildkitBackend) String() string {
	return "buildkit-backend"
}

func (backend *BuildkitBackend) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error) {
	return "", fmt.Errorf("build dockerfile: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	return "", fmt.Errorf("build dockerfile stage: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error) {
	return "", fmt.Errorf("build stapel stage: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	return fmt.Errorf("tag image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	return fmt.Errorf("push image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	return fmt.Errorf("pull image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	return fmt.Errorf("remove image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Rm(ctx context.Context, name string, opts RmOpts) error {
	return fmt.Errorf("remove container: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error {
	return fmt.Errorf("post manifest: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	return nil, fmt.Errorf("get image info: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error) {
	return image.ImagesList{}, nil
}

func (backend *BuildkitBackend) Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error) {
	return image.ContainerList{}, nil
}

func (backend *BuildkitBackend) PruneImages(ctx context.Context, options prune.Options) (prune.Report, error) {
	return prune.Report{}, ErrUnsupportedFeature
}

func (backend *BuildkitBackend) PruneVolumes(ctx context.Context, options prune.Options) (prune.Report, error) {
	return prune.Report{}, ErrUnsupportedFeature
}

func (backend *BuildkitBackend) RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error {
	return fmt.Errorf("remove host dirs: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	return fmt.Errorf("refresh image object: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	return fmt.Errorf("pull image from registry: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	return fmt.Errorf("rename image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	return fmt.Errorf("remove image: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) error {
	return fmt.Errorf("tag image by name: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("save image to stream: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) LoadImageFromStream(ctx context.Context, input io.Reader) (string, error) {
	return "", fmt.Errorf("load image from stream: %w", ErrUnsupportedFeature)
}
