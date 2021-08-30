package build

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
)

type BaseImageType string

const (
	ImageFromRegistryAsBaseImage BaseImageType = "ImageFromRegistryBaseImage"
	StageAsBaseImage             BaseImageType = "StageBaseImage"
)

type Image struct {
	name string

	baseImageName      string
	baseImageImageName string
	baseImageRepoId    string

	stages            []stage.Interface
	lastNonEmptyStage stage.Interface
	contentDigest     string
	isArtifact        bool
	isDockerfileImage bool

	baseImageType    BaseImageType
	stageAsBaseImage stage.Interface
	baseImage        *container_runtime.StageImage
}

func (i *Image) LogName() string {
	return logging.ImageLogName(i.name, i.isArtifact)
}

func (i *Image) LogDetailedName() string {
	return logging.ImageLogProcessName(i.name, i.isArtifact)
}

func (i *Image) LogProcessStyle() color.Style {
	return ImageLogProcessStyle(i.isArtifact)
}

func (i *Image) LogTagStyle() color.Style {
	return ImageLogTagStyle(i.isArtifact)
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
	return i.GetLastNonEmptyStage().GetImage().GetStageDescription().Info.Tag
}

func (i *Image) GetName() string {
	return i.name
}

func (i *Image) GetLogName() string {
	return i.LogName()
}

func (i *Image) SetupBaseImage(c *Conveyor) {
	if i.baseImageImageName != "" {
		i.baseImageType = StageAsBaseImage
		i.stageAsBaseImage = c.GetImage(i.baseImageImageName).GetLastNonEmptyStage()
		i.baseImage = c.GetOrCreateStageImage(nil, i.stageAsBaseImage.GetImage().Name())
	} else {
		i.baseImageType = ImageFromRegistryAsBaseImage
		i.baseImage = c.GetOrCreateStageImage(nil, i.baseImageName)
	}
}

func (i *Image) GetBaseImage() *container_runtime.StageImage {
	return i.baseImage
}

func (i *Image) FetchBaseImage(ctx context.Context, c *Conveyor) error {
	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		containerRuntime := c.ContainerRuntime.(*container_runtime.LocalDockerServerRuntime)

		if inspect, err := containerRuntime.GetImageInspect(ctx, i.baseImage.Name()); err != nil {
			return fmt.Errorf("unable to inspect local image %s: %s", i.baseImage.Name(), err)
		} else if inspect != nil {
			// TODO: do not use container_runtime.StageImage for base image
			i.baseImage.SetStageDescription(&image.StageDescription{
				StageID: nil, // this is not a stage actually, TODO
				Info:    image.NewInfoFromInspect(i.baseImage.Name(), inspect),
			})

			baseImageRepoId, err := i.getFromBaseImageIdFromRegistry(ctx, c, i.baseImage.Name())
			if baseImageRepoId == inspect.ID || err != nil {
				if err != nil {
					logboek.Context(ctx).Warn().LogF("WARNING: cannot get base image id (%s): %s\n", i.baseImage.Name(), err)
					logboek.Context(ctx).Warn().LogF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
					logboek.Context(ctx).Warn().LogOptionalLn()
				}

				return nil
			}
		}

		if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", i.baseImage.Name()).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				return c.ContainerRuntime.PullImageFromRegistry(ctx, &container_runtime.DockerImage{Image: i.baseImage})
			}); err != nil {
			return err
		}

		if inspect, err := containerRuntime.GetImageInspect(ctx, i.baseImage.Name()); err != nil {
			return fmt.Errorf("unable to inspect local image %s: %s", i.baseImage.Name(), err)
		} else if inspect == nil {
			return fmt.Errorf("unable to inspect local image %s after successful pull: image is not exists", i.baseImage.Name())
		} else {
			i.baseImage.SetStageDescription(&image.StageDescription{
				StageID: nil, // this is not a stage actually, TODO
				Info:    image.NewInfoFromInspect(i.baseImage.Name(), inspect),
			})
		}
	case StageAsBaseImage:
		if err := c.ContainerRuntime.RefreshImageObject(ctx, &container_runtime.DockerImage{Image: i.baseImage}); err != nil {
			return err
		}
		if err := c.StorageManager.FetchStage(ctx, c.ContainerRuntime, i.stageAsBaseImage); err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}

	return nil
}

func (i *Image) getFromBaseImageIdFromRegistry(ctx context.Context, c *Conveyor, baseImageName string) (string, error) {
	c.getServiceRWMutex("baseImagesRepoIdsCache" + baseImageName).Lock()
	defer c.getServiceRWMutex("baseImagesRepoIdsCache" + baseImageName).Unlock()

	if i.baseImageRepoId != "" {
		return i.baseImageRepoId, nil
	} else if c.IsBaseImagesRepoIdsCacheExist(baseImageName) {
		i.baseImageRepoId = c.GetBaseImagesRepoIdsCache(baseImageName)
		return i.baseImageRepoId, nil
	} else if c.IsBaseImagesRepoErrCacheExist(baseImageName) {
		return "", c.GetBaseImagesRepoErrCache(baseImageName)
	}

	var fetchedBaseRepoImage *image.Info
	processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", baseImageName)
	if err := logboek.Context(ctx).Info().LogProcessInline(processMsg).DoError(func() error {
		var fetchImageIdErr error
		fetchedBaseRepoImage, fetchImageIdErr = docker_registry.API().GetRepoImage(ctx, baseImageName)
		if fetchImageIdErr != nil {
			c.SetBaseImagesRepoErrCache(baseImageName, fetchImageIdErr)
			return fmt.Errorf("can not get base image id from registry (%s): %s", baseImageName, fetchImageIdErr)
		}

		return nil
	}); err != nil {
		return "", err
	}

	i.baseImageRepoId = fetchedBaseRepoImage.ID
	c.SetBaseImagesRepoIdsCache(baseImageName, i.baseImageRepoId)

	return i.baseImageRepoId, nil
}
