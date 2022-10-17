package image

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage/manager"
)

type BaseImageType string

const (
	ImageFromRegistryAsBaseImage BaseImageType = "ImageFromRegistryBaseImage"
	StageAsBaseImage             BaseImageType = "StageBaseImage"
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
}

type ImageOptions struct {
	CommonImageOptions
	IsArtifact, IsDockerfileImage bool
	DockerfileImageConfig         *config.ImageFromDockerfile

	BaseImageReference   string
	BaseImageName        string
	FetchLatestBaseImage bool
}

func NewImage(ctx context.Context, name string, baseImageType BaseImageType, opts ImageOptions) (*Image, error) {
	switch baseImageType {
	case NoBaseImage, ImageFromRegistryAsBaseImage, StageAsBaseImage:
	default:
		panic(fmt.Sprintf("unknown opts.BaseImageType %q", baseImageType))
	}

	i := &Image{
		Name:                  name,
		CommonImageOptions:    opts.CommonImageOptions,
		IsArtifact:            opts.IsArtifact,
		IsDockerfileImage:     opts.IsDockerfileImage,
		DockerfileImageConfig: opts.DockerfileImageConfig,

		baseImageType:      baseImageType,
		baseImageReference: opts.BaseImageReference,
		baseImageName:      opts.BaseImageName,
	}

	if opts.FetchLatestBaseImage {
		if _, err := i.getFromBaseImageIdFromRegistry(ctx, i.baseImageReference); err != nil {
			return nil, fmt.Errorf("error fetching base image id from registry: %w", err)
		}
	}

	return i, nil
}

type Image struct {
	CommonImageOptions

	IsArtifact            bool
	IsDockerfileImage     bool
	Name                  string
	DockerfileImageConfig *config.ImageFromDockerfile

	stages            []stage.Interface
	lastNonEmptyStage stage.Interface
	contentDigest     string
	rebuilt           bool

	baseImageType      BaseImageType
	baseImageReference string
	baseImageName      string

	baseImageRepoId  string
	baseStageImage   *stage.StageImage
	stageAsBaseImage stage.Interface
}

func (i *Image) LogName() string {
	return logging.ImageLogName(i.Name, i.IsArtifact)
}

func (i *Image) LogDetailedName() string {
	return logging.ImageLogProcessName(i.Name, i.IsArtifact)
}

func (i *Image) LogProcessStyle() color.Style {
	return ImageLogProcessStyle(i.IsArtifact)
}

func (i *Image) LogTagStyle() color.Style {
	return ImageLogTagStyle(i.IsArtifact)
}

func ImageLogProcessStyle(isArtifact bool) color.Style {
	return imageDefaultStyle(isArtifact)
}

func ImageLogTagStyle(isArtifact bool) color.Style {
	return imageDefaultStyle(isArtifact)
}

func imageDefaultStyle(isArtifact bool) color.Style {
	var colors []color.Color
	if isArtifact {
		colors = []color.Color{color.FgCyan, color.Bold}
	} else {
		colors = []color.Color{color.FgYellow, color.Bold}
	}

	return color.New(colors...)
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

func (i *Image) SetupBaseImage() {
	switch i.baseImageType {
	case StageAsBaseImage:
		i.stageAsBaseImage = i.Conveyor.GetImage(i.baseImageName).GetLastNonEmptyStage()
		i.baseImageReference = i.stageAsBaseImage.GetStageImage().Image.Name()
		i.baseStageImage = i.stageAsBaseImage.GetStageImage()
	case ImageFromRegistryAsBaseImage:
		i.baseStageImage = i.Conveyor.GetOrCreateStageImage(i.baseImageReference, nil, nil, i)
	case NoBaseImage:
	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}
}

// TODO(staged-dockerfile): this is only for compatibility with stapel-builder logic, and this should be unified with new staged-dockerfile logic
func (i *Image) GetBaseStageImage() *stage.StageImage {
	return i.baseStageImage
}

func (i *Image) GetBaseImageReference() string {
	return i.baseImageReference
}

func (i *Image) FetchBaseImage(ctx context.Context) error {
	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		if info, err := i.ContainerBackend.GetImageInfo(ctx, i.baseStageImage.Image.Name(), container_backend.GetImageInfoOpts{}); err != nil {
			return fmt.Errorf("unable to inspect local image %s: %w", i.baseStageImage.Image.Name(), err)
		} else if info != nil {
			// TODO: do not use container_backend.LegacyStageImage for base image
			i.baseStageImage.Image.SetStageDescription(&image.StageDescription{
				StageID: nil, // this is not a stage actually, TODO
				Info:    info,
			})

			baseImageRepoId, err := i.getFromBaseImageIdFromRegistry(ctx, i.baseStageImage.Image.Name())
			if baseImageRepoId == info.ID || err != nil {
				if err != nil {
					logboek.Context(ctx).Warn().LogF("WARNING: cannot get base image id (%s): %s\n", i.baseStageImage.Image.Name(), err)
					logboek.Context(ctx).Warn().LogF("WARNING: using existing image %s without pull\n", i.baseStageImage.Image.Name())
					logboek.Context(ctx).Warn().LogOptionalLn()
				}

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

func (i *Image) getFromBaseImageIdFromRegistry(ctx context.Context, reference string) (string, error) {
	i.Conveyor.GetServiceRWMutex("baseImagesRepoIdsCache" + reference).Lock()
	defer i.Conveyor.GetServiceRWMutex("baseImagesRepoIdsCache" + reference).Unlock()

	switch {
	case i.baseImageRepoId != "":
		return i.baseImageRepoId, nil
	case i.Conveyor.IsBaseImagesRepoIdsCacheExist(reference):
		i.baseImageRepoId = i.Conveyor.GetBaseImagesRepoIdsCache(reference)
		return i.baseImageRepoId, nil
	case i.Conveyor.IsBaseImagesRepoErrCacheExist(reference):
		return "", i.Conveyor.GetBaseImagesRepoErrCache(reference)
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
		return "", err
	}

	i.baseImageRepoId = fetchedBaseRepoImage.ID
	i.Conveyor.SetBaseImagesRepoIdsCache(reference, i.baseImageRepoId)

	return i.baseImageRepoId, nil
}
