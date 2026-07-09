package container_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/containerd/containerd/platforms"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/session"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tonistiigi/fsutil"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/buildkit"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
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

	mu             sync.Mutex
	repo           string
	dockerRegistry docker_registry.Interface
	client         *bkclient.Client
}

type BuildkitBackendOptions struct {
	DefaultPlatform string
	DockerConfigDir string
}

func NewBuildkitBackend(host string, opts BuildkitBackendOptions) *BuildkitBackend {
	return &BuildkitBackend{host: host, BuildkitBackendOptions: opts}
}

func (backend *BuildkitBackend) getClient(ctx context.Context) (*bkclient.Client, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if backend.client != nil {
		return backend.client, nil
	}

	client, err := buildkit.NewClient(ctx, backend.host)
	if err != nil {
		return nil, err
	}
	backend.client = client
	return backend.client, nil
}

func (backend *BuildkitBackend) parsePlatform(targetPlatform string) (*ocispecs.Platform, error) {
	if targetPlatform == "" {
		targetPlatform = backend.GetDefaultPlatform()
	}
	platform, err := platforms.Parse(targetPlatform)
	if err != nil {
		return nil, fmt.Errorf("parse platform %q: %w", targetPlatform, err)
	}
	return &platform, nil
}

func (backend *BuildkitBackend) getSessionAttachables(ssh string, secrets []string) ([]session.Attachable, error) {
	sshSpec := ssh
	if sshSpec == "" && ssh_agent.SSHAuthSock != "" {
		sshSpec = "default"
	}

	sshAgentSocks, err := buildkit.ParseSSHSpec(sshSpec, ssh_agent.SSHAuthSock)
	if err != nil {
		return nil, fmt.Errorf("parse ssh spec: %w", err)
	}

	secretSources, err := buildkit.ParseSecretSpecs(secrets)
	if err != nil {
		return nil, fmt.Errorf("parse secret specs: %w", err)
	}

	return buildkit.SessionAttachables(buildkit.SessionAttachablesOptions{
		DockerConfigDir: backend.DockerConfigDir,
		SSHAgentSocks:   sshAgentSocks,
		Secrets:         secretSources,
	})
}

func parseKeyValuePairs(pairs []string, what string) (map[string]string, error) {
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid %s %q given, expected string in the key=value format", what, pair)
		}
		res[parts[0]] = parts[1]
	}
	return res, nil
}

func parseExtraHosts(addHost []string) ([]llb.HostIP, error) {
	var res []llb.HostIP
	for _, h := range addHost {
		host, ip, ok := strings.Cut(h, ":")
		if !ok {
			return nil, fmt.Errorf("invalid add-host %q given, expected string in the host:ip format", h)
		}
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("invalid add-host %q given: invalid ip %q", h, ip)
		}
		res = append(res, llb.HostIP{Host: host, IP: parsedIP})
	}
	return res, nil
}

func (backend *BuildkitBackend) SetStagesStorage(repo string, dockerRegistry docker_registry.Interface) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	backend.repo = repo
	backend.dockerRegistry = dockerRegistry
}

func (backend *BuildkitBackend) getStagesStorageRepo() (string, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if backend.repo == "" {
		return "", fmt.Errorf("stages storage repo is not set: --repo is required when using buildkit backend")
	}
	return backend.repo, nil
}

func (backend *BuildkitBackend) getDockerRegistry() (docker_registry.Interface, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if backend.dockerRegistry == nil {
		return nil, fmt.Errorf("stages storage docker registry is not set: --repo is required when using buildkit backend")
	}
	return backend.dockerRegistry, nil
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
	if opts.BuildContextArchive == nil {
		panic(fmt.Sprintf("BuildContextArchive can't be nil: %+v", opts))
	}

	repo, err := backend.getStagesStorageRepo()
	if err != nil {
		return "", err
	}

	cl, err := backend.getClient(ctx)
	if err != nil {
		return "", err
	}

	contextDir, err := opts.BuildContextArchive.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to extract build context: %w", err)
	}

	buildArgs, err := parseKeyValuePairs(opts.BuildArgs, "build argument")
	if err != nil {
		return "", err
	}

	labels, err := parseKeyValuePairs(opts.Labels, "label")
	if err != nil {
		return "", err
	}

	extraHosts, err := parseExtraHosts(opts.AddHost)
	if err != nil {
		return "", err
	}

	netMode, err := buildkit.ParseNetMode(opts.Network)
	if err != nil {
		return "", err
	}

	platform, err := backend.parsePlatform(opts.TargetPlatform)
	if err != nil {
		return "", err
	}

	contextState := llb.Local(dockerui.DefaultLocalNameContext)

	convertResult, err := dockerfile2llb.Dockerfile2LLB(ctx, dockerfile, dockerfile2llb.ConvertOpt{
		Config: dockerui.Config{
			BuildArgs:   buildArgs,
			Labels:      labels,
			Target:      opts.Target,
			ExtraHosts:  extraHosts,
			NetworkMode: netMode,
		},
		MainContext:    &contextState,
		TargetPlatform: platform,
		MetaResolver:   buildkit.NewImageMetaResolver(platform),
	})
	if err != nil {
		return "", fmt.Errorf("convert dockerfile to llb: %w", err)
	}

	def, err := convertResult.State.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("marshal llb state: %w", err)
	}

	imageConfig, err := json.Marshal(convertResult.Image)
	if err != nil {
		return "", fmt.Errorf("marshal image config: %w", err)
	}

	attachables, err := backend.getSessionAttachables(opts.SSH, opts.Secrets)
	if err != nil {
		return "", err
	}

	contextFS, err := fsutil.NewFS(contextDir)
	if err != nil {
		return "", fmt.Errorf("create fs for build context %q: %w", contextDir, err)
	}

	builtID, err := buildkit.Solve(ctx, cl, def, buildkit.SolveOptions{
		Repo:        repo,
		ImageConfig: imageConfig,
		LocalMounts: map[string]fsutil.FS{dockerui.DefaultLocalNameContext: contextFS},
		Session:     attachables,
	})
	if err != nil {
		return "", fmt.Errorf("build dockerfile: %w", err)
	}

	return builtID, nil
}

// Tag creates a registry-side tag pointing to the already-pushed image (built images are
// pushed by digest during Solve), without re-uploading layers.
func (backend *BuildkitBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	dockerRegistry, err := backend.getDockerRegistry()
	if err != nil {
		return err
	}
	if err := dockerRegistry.CopyImage(ctx, ref, newRef, docker_registry.CopyImageOptions{}); err != nil {
		return fmt.Errorf("tag image %q as %q in registry: %w", ref, newRef, err)
	}
	return nil
}

// Push is a no-op: images are pushed by digest during Solve and tagged registry-side by Tag.
func (backend *BuildkitBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	return nil
}

func (backend *BuildkitBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	return fmt.Errorf("pull image to local store (there is no local image store with buildkit backend): %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	dockerRegistry, err := backend.getDockerRegistry()
	if err != nil {
		return err
	}

	info, err := dockerRegistry.TryGetRepoImage(ctx, ref)
	if err != nil {
		return fmt.Errorf("get image %q from registry: %w", ref, err)
	}
	if info == nil {
		return nil
	}

	if err := dockerRegistry.DeleteRepoImage(ctx, info); err != nil {
		return fmt.Errorf("delete image %q from registry: %w", ref, err)
	}
	return nil
}

func (backend *BuildkitBackend) Rm(ctx context.Context, name string, opts RmOpts) error {
	return nil
}

func (backend *BuildkitBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error {
	return fmt.Errorf("post manifest (only repo stages storage is supported by buildkit backend and it posts manifests registry-side): %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	dockerRegistry, err := backend.getDockerRegistry()
	if err != nil {
		return nil, err
	}

	info, err := dockerRegistry.TryGetRepoImage(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("get image %q from registry: %w", ref, err)
	}
	return info, nil
}

func (backend *BuildkitBackend) Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error) {
	return image.ImagesList{}, nil
}

func (backend *BuildkitBackend) Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error) {
	return image.ContainerList{}, nil
}

// PruneImages prunes the buildkitd build cache: there is no local image store to prune.
func (backend *BuildkitBackend) PruneImages(ctx context.Context, options prune.Options) (prune.Report, error) {
	cl, err := backend.getClient(ctx)
	if err != nil {
		return prune.Report{}, err
	}

	ch := make(chan bkclient.UsageInfo)
	report := prune.Report{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for usage := range ch {
			report.ItemsDeleted = append(report.ItemsDeleted, usage.ID)
			report.SpaceReclaimed += uint64(usage.Size)
		}
	}()

	if err := cl.Prune(ctx, ch); err != nil {
		return prune.Report{}, fmt.Errorf("prune buildkitd cache: %w", err)
	}
	<-done

	return report, nil
}

func (backend *BuildkitBackend) PruneVolumes(ctx context.Context, options prune.Options) (prune.Report, error) {
	return prune.Report{}, ErrUnsupportedFeature
}

// RemoveHostDirs removes werf-owned host dirs directly: unlike docker/buildah there is no
// container or chroot reexec to elevate privileges, and the buildkit path only produces
// werf-owned files on the host.
func (backend *BuildkitBackend) RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove host dir %q: %w", dir, err)
		}
	}
	return nil
}

func (backend *BuildkitBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()})
	if err != nil {
		return err
	}
	img.SetInfo(info)
	return nil
}

// PullImageFromRegistry refreshes the image info from the registry: there is no local image
// store to pull into, builds fetch base images inside buildkitd.
func (backend *BuildkitBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()})
	if err != nil {
		return fmt.Errorf("unable to get info of image %s: %w", img.Name(), err)
	}
	if info == nil {
		return fmt.Errorf("image %s not found in registry", img.Name())
	}
	img.SetInfo(info)
	return nil
}

// RenameImage copies the image to the new reference registry-side (registries have no rename).
func (backend *BuildkitBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	dockerRegistry, err := backend.getDockerRegistry()
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := dockerRegistry.CopyImage(ctx, img.Name(), newImageName, docker_registry.CopyImageOptions{}); err != nil {
			return fmt.Errorf("unable to copy image %s to %s in registry: %w", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			return backend.Rmi(ctx, img.Name(), RmiOpts{CommonOpts: CommonOpts{TargetPlatform: img.GetTargetPlatform()}})
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if err := backend.RefreshImageObject(ctx, img); err != nil {
		return err
	}

	if stageDesc := img.GetStageDesc(); stageDesc != nil {
		repository, tag := image.ParseRepositoryAndTag(newImageName)
		stageDesc.Info.Name = newImageName
		stageDesc.Info.Repository = repository
		stageDesc.Info.Tag = tag
	}

	return nil
}

func (backend *BuildkitBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	return logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		return backend.Rmi(ctx, img.Name(), RmiOpts{
			CommonOpts: CommonOpts{TargetPlatform: img.GetTargetPlatform()},
		})
	})
}

func (backend *BuildkitBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) error {
	if img.BuiltID() != "" {
		if err := backend.Tag(ctx, img.BuiltID(), img.Name(), TagOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
			return fmt.Errorf("unable to tag image %s: %w", img.Name(), err)
		}
		return nil
	}
	return backend.RefreshImageObject(ctx, img)
}

func (backend *BuildkitBackend) SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("save image to stream: %w", ErrUnsupportedFeature)
}

func (backend *BuildkitBackend) LoadImageFromStream(ctx context.Context, input io.Reader) (string, error) {
	return "", fmt.Errorf("load image from stream: %w", ErrUnsupportedFeature)
}
