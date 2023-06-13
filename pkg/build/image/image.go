package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/gookit/color"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/werf"
)

type BaseImageType string

const (
	ImageFromRegistryAsBaseImage BaseImageType = "ImageFromRegistryAsBaseImage"
	StageAsBaseImage             BaseImageType = "StageAsBaseImage"
	NoBaseImage                  BaseImageType = "NoBaseImage"
)

type CommonImageOptions struct {
	Conveyor           Conveyor
	GiterminismManager giterminism_manager.Interface
	ContainerBackend   container_backend.ContainerBackend
	StorageManager     manager.StorageManagerInterface
	ProjectDir         string
	ProjectName        string
	ContainerWerfDir   string
	TmpDir             string

	ForceTargetPlatformLogging bool
}

type ImageOptions struct {
	CommonImageOptions
	IsArtifact, IsDockerfileImage, IsDockerfileTargetStage bool
	DockerfileImageConfig                                  *config.ImageFromDockerfile

	BaseImageReference        string
	BaseImageName             string
	FetchLatestBaseImage      bool
	DockerfileExpanderFactory dockerfile.ExpanderFactory
}

func NewImage(ctx context.Context, targetPlatform, name string, baseImageType BaseImageType, opts ImageOptions) (*Image, error) {
	switch baseImageType {
	case NoBaseImage, ImageFromRegistryAsBaseImage, StageAsBaseImage:
	default:
		panic(fmt.Sprintf("unknown opts.BaseImageType %q", baseImageType))
	}

	if targetPlatform == "" {
		panic("assertion: targetPlatform cannot be empty")
	}

	i := &Image{
		Name:                    name,
		CommonImageOptions:      opts.CommonImageOptions,
		IsArtifact:              opts.IsArtifact,
		IsDockerfileImage:       opts.IsDockerfileImage,
		IsDockerfileTargetStage: opts.IsDockerfileTargetStage,
		DockerfileImageConfig:   opts.DockerfileImageConfig,
		TargetPlatform:          targetPlatform,

		baseImageType:             baseImageType,
		baseImageReference:        opts.BaseImageReference,
		baseImageName:             opts.BaseImageName,
		dockerfileExpanderFactory: opts.DockerfileExpanderFactory,
	}

	if opts.FetchLatestBaseImage {
		if err := i.setupBaseImageRepoDigest(ctx, i.baseImageReference); err != nil {
			return nil, fmt.Errorf("error fetching base image id from registry: %w", err)
		}
	}

	return i, nil
}

type Image struct {
	CommonImageOptions

	IsArtifact              bool
	IsDockerfileImage       bool
	IsDockerfileTargetStage bool
	Name                    string
	DockerfileImageConfig   *config.ImageFromDockerfile
	TargetPlatform          string

	stages            []stage.Interface
	lastNonEmptyStage stage.Interface
	contentDigest     string
	rebuilt           bool

	baseImageType             BaseImageType
	baseImageReference        string
	baseImageName             string
	dockerfileExpanderFactory dockerfile.ExpanderFactory

	// NOTICE: baseImageRepoId is a legacy field, better use Digest instead everywhere
	baseImageRepoId     string
	baseImageRepoDigest string

	baseStageImage   *stage.StageImage
	stageAsBaseImage stage.Interface
}

func (i *Image) LogName() string {
	return logging.ImageLogName(i.Name, i.IsArtifact)
}

func (i *Image) ShouldLogPlatform() bool {
	return i.ForceTargetPlatformLogging || i.TargetPlatform != i.ContainerBackend.GetRuntimePlatform()
}

func (i *Image) LogDetailedName() string {
	var targetPlatformForLog string
	if i.ShouldLogPlatform() {
		targetPlatformForLog = i.TargetPlatform
	}
	return logging.ImageLogProcessName(i.Name, i.IsArtifact, targetPlatformForLog)
}

func (i *Image) LogProcessStyle() color.Style {
	return ImageLogProcessStyle(i.IsArtifact)
}

func (i *Image) LogTagStyle() color.Style {
	return ImageLogTagStyle(i.IsArtifact)
}

func ImageLogProcessStyle(isArtifact bool) color.Style {
	return logging.ImageDefaultStyle(isArtifact)
}

func ImageLogTagStyle(isArtifact bool) color.Style {
	return logging.ImageDefaultStyle(isArtifact)
}

func (i *Image) IsFinal() bool {
	if i.IsArtifact {
		return false
	}
	if i.IsDockerfileImage && !i.IsDockerfileTargetStage {
		return false
	}
	return true
}

func (i *Image) SetStages(stages []stage.Interface) {
	i.stages = stages
}

func (i *Image) GetStages() []stage.Interface {
	return i.stages
}

func (i *Image) SetLastNonEmptyStage(stg stage.Interface) {
	i.lastNonEmptyStage = stg
}

func (i *Image) GetLastNonEmptyStage() stage.Interface {
	return i.lastNonEmptyStage
}

func (i *Image) SetContentDigest(digest string) {
	i.contentDigest = digest
}

func (i *Image) GetContentDigest() string {
	return i.contentDigest
}

func (i *Image) GetStage(name stage.StageName) stage.Interface {
	for _, s := range i.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
}

func (i *Image) GetStageID() string {
	return i.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription().Info.Tag
}

func (i *Image) UsesBuildContext() bool {
	for _, stg := range i.GetStages() {
		if stg.UsesBuildContext() {
			return true
		}
	}

	return false
}

func (i *Image) GetName() string {
	return i.Name
}

func (i *Image) GetLogName() string {
	return i.LogName()
}

func (i *Image) SetRebuilt(rebuilt bool) {
	i.rebuilt = rebuilt
}

func (i *Image) GetRebuilt() bool {
	return i.rebuilt
}

func (i *Image) ExpandDependencies(ctx context.Context, baseEnv map[string]string) error {
	for _, stg := range i.stages {
		if err := stg.ExpandDependencies(ctx, i.Conveyor, baseEnv); err != nil {
			return fmt.Errorf("unable to expand dependencies for stage %q: %w", stg.Name(), err)
		}
	}
	return nil
}

func isUnsupportedMediaTypeError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unsupported MediaType")
}

func (i *Image) SetupBaseImage(ctx context.Context, storageManager manager.StorageManagerInterface, storageOpts manager.StorageOptions) error {
	logboek.Context(ctx).Debug().LogF(" -- SetupBaseImage for %q\n", i.Name)

	switch i.baseImageType {
	case StageAsBaseImage:
		i.stageAsBaseImage = i.Conveyor.GetImage(i.TargetPlatform, i.baseImageName).GetLastNonEmptyStage()
		i.baseImageReference = i.stageAsBaseImage.GetStageImage().Image.Name()
		i.baseStageImage = i.stageAsBaseImage.GetStageImage()

	case ImageFromRegistryAsBaseImage:
		if i.IsDockerfileImage && i.dockerfileExpanderFactory != nil {
			dependenciesArgs := stage.ResolveDependenciesArgs(i.TargetPlatform, i.DockerfileImageConfig.Dependencies, i.Conveyor)
			ref, err := i.dockerfileExpanderFactory.GetExpander(dockerfile.ExpandOptions{SkipUnsetEnv: false}).ProcessWordWithMap(i.baseImageReference, dependenciesArgs)
			if err != nil {
				return fmt.Errorf("unable to expand dockerfile base image reference %q: %w", i.baseImageReference, err)
			}
			i.baseImageReference = ref
		}

		i.baseStageImage = i.Conveyor.GetOrCreateStageImage(i.baseImageReference, nil, nil, i)

		if i.IsDockerfileImage && i.DockerfileImageConfig.Staged {
			if werf.GetStagedDockerfileVersion() == werf.StagedDockerfileV1 {
				var info *image.Info

				if i.baseImageReference != "scratch" {
					var err error
					info, err = storageManager.GetImageInfo(ctx, i.baseImageReference, storageOpts)
					if isUnsupportedMediaTypeError(err) {
						if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", i.baseStageImage.Image.Name()).
							Options(func(options types.LogProcessOptionsInterface) {
								options.Style(style.Highlight())
							}).
							DoError(func() error {
								return i.ContainerBackend.PullImageFromRegistry(ctx, i.baseStageImage.Image)
							}); err != nil {
							return err
						}

						info, err = storageManager.GetImageInfo(ctx, i.baseImageReference, storageOpts)
						if err != nil {
							return fmt.Errorf("unable to get base image %q manifest: %w", i.baseImageReference, err)
						}
					} else if err != nil {
						return fmt.Errorf("unable to get base image %q manifest: %w", i.baseImageReference, err)
					}
				} else {
					info = &image.Info{
						Name: i.baseImageReference,
						Env:  nil,
					}
				}

				i.baseStageImage.Image.SetStageDescription(&image.StageDescription{
					StageID: nil, // this is not a stage actually, TODO
					Info:    info,
				})
			}
		}
	case NoBaseImage:

	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}

	if i.IsDockerfileImage && i.DockerfileImageConfig.Staged {
		if werf.GetStagedDockerfileVersion() == werf.StagedDockerfileV1 {
			switch i.baseImageType {
			case StageAsBaseImage, ImageFromRegistryAsBaseImage:
				if err := i.ExpandDependencies(ctx, EnvToMap(i.baseStageImage.Image.GetStageDescription().Info.Env)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// TODO(staged-dockerfile): this is only for compatibility with stapel-builder logic, and this should be unified with new staged-dockerfile logic
func (i *Image) GetBaseStageImage() *stage.StageImage {
	return i.baseStageImage
}

func (i *Image) GetBaseImageReference() string {
	return i.baseImageReference
}

func (i *Image) GetBaseImageRepoDigest() string {
	return i.baseImageRepoDigest
}

func (i *Image) FetchBaseImage(ctx context.Context) error {
	logboek.Context(ctx).Debug().LogF(" -- FetchBaseImage for %q\n", i.Name)

	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		if i.baseStageImage.Image.Name() == "scratch" {
			return nil
		}

		// TODO: Refactor, move manifest fetching into SetupBaseImage, only pull image in FetchBaseImage method

		if info, err := i.ContainerBackend.GetImageInfo(ctx, i.baseStageImage.Image.Name(), container_backend.GetImageInfoOpts{}); err != nil {
			return fmt.Errorf("unable to inspect local image %s: %w", i.baseStageImage.Image.Name(), err)
		} else if info != nil {
			logboek.Context(ctx).Debug().LogF("GetImageInfo of %q -> %#v\n", i.baseStageImage.Image.Name(), info)

			// TODO: do not use container_backend.LegacyStageImage for base image
			i.baseStageImage.Image.SetStageDescription(&image.StageDescription{
				StageID: nil, // this is not a stage actually, TODO
				Info:    info,
			})

			err = i.setupBaseImageRepoDigest(ctx, i.baseStageImage.Image.Name())
			if (i.baseImageRepoDigest != "" && i.baseImageRepoDigest == info.RepoDigest) || (err != nil && !isUnsupportedMediaTypeError(err)) {
				if err != nil {
					logboek.Context(ctx).Warn().LogF("WARNING: cannot get base image id (%s): %s\n", i.baseStageImage.Image.Name(), err)
					logboek.Context(ctx).Warn().LogF("WARNING: using existing image %s without pull\n", i.baseStageImage.Image.Name())
					logboek.Context(ctx).Warn().LogOptionalLn()
				} else {
					logboek.Context(ctx).Info().LogF("No pull needed for base image %s of image %q: image by digest %s is up to date\n", i.baseImageReference, i.Name, i.baseImageRepoDigest)
				}
				// No image pull
				return nil
			}
		}

		if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", i.baseStageImage.Image.Name()).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				return i.ContainerBackend.PullImageFromRegistry(ctx, i.baseStageImage.Image)
			}); err != nil {
			return err
		}

		info, err := i.ContainerBackend.GetImageInfo(ctx, i.baseStageImage.Image.Name(), container_backend.GetImageInfoOpts{})
		if err != nil {
			return fmt.Errorf("unable to inspect local image %s: %w", i.baseStageImage.Image.Name(), err)
		}

		if info == nil {
			return fmt.Errorf("unable to inspect local image %s after successful pull: image is not exists", i.baseStageImage.Image.Name())
		}

		i.baseStageImage.Image.SetStageDescription(&image.StageDescription{
			StageID: nil, // this is not a stage actually, TODO
			Info:    info,
		})

		return nil
	case StageAsBaseImage:
		return i.StorageManager.FetchStage(ctx, i.ContainerBackend, i.stageAsBaseImage)

	case NoBaseImage:
		return nil

	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}
}

func packRepoIDAndDigest(repoID, digest string) string {
	return fmt.Sprintf("%s/%s", repoID, digest)
}

func unpackRepoIDAndDigest(packed string) (string, string) {
	parts := strings.SplitN(packed, "/", 2)
	return parts[0], parts[1]
}

func (i *Image) setupBaseImageRepoDigest(ctx context.Context, reference string) error {
	i.Conveyor.GetServiceRWMutex("baseImagesRepoIdsCache" + reference).Lock()
	defer i.Conveyor.GetServiceRWMutex("baseImagesRepoIdsCache" + reference).Unlock()

	switch {
	case i.baseImageRepoId != "":
		return nil
	case i.Conveyor.IsBaseImagesRepoIdsCacheExist(reference):
		i.baseImageRepoId, i.baseImageRepoDigest = unpackRepoIDAndDigest(i.Conveyor.GetBaseImagesRepoIdsCache(reference))
		return nil
	case i.Conveyor.IsBaseImagesRepoErrCacheExist(reference):
		return i.Conveyor.GetBaseImagesRepoErrCache(reference)
	}

	var fetchedBaseRepoImage *image.Info
	processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", reference)
	if err := logboek.Context(ctx).Info().LogProcessInline(processMsg).DoError(func() error {
		var fetchImageIdErr error
		fetchedBaseRepoImage, fetchImageIdErr = docker_registry.API().GetRepoImage(ctx, reference)
		if fetchImageIdErr != nil {
			i.Conveyor.SetBaseImagesRepoErrCache(reference, fetchImageIdErr)
			return fmt.Errorf("can not get base image id from registry (%s): %w", reference, fetchImageIdErr)
		}

		return nil
	}); err != nil {
		return err
	}

	i.baseImageRepoId = fetchedBaseRepoImage.ID
	i.baseImageRepoDigest = fetchedBaseRepoImage.RepoDigest
	i.Conveyor.SetBaseImagesRepoIdsCache(reference, packRepoIDAndDigest(i.baseImageRepoId, i.baseImageRepoDigest))

	return nil
}

func EnvToMap(env []string) map[string]string {
	res := make(map[string]string)
	for _, kv := range env {
		k, v := parseKeyValue(kv)
		res[k] = v
	}
	return res
}

func parseKeyValue(env string) (string, string) {
	parts := strings.SplitN(env, "=", 2)
	v := ""
	if len(parts) > 1 {
		v = parts[1]
	}

	return parts[0], v
}
