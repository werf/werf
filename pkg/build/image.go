package build

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend"
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
	rebuilt           bool

	baseImageType    BaseImageType
	stageAsBaseImage stage.Interface
	baseImage        *stage.StageImage
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
	return i.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription().Info.Tag
}

func (i *Image) GetName() string {
	return i.name
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

func (i *Image) SetupBaseImage(c *Conveyor) {
	if i.baseImageImageName != "" {
		i.baseImageType = StageAsBaseImage
		i.stageAsBaseImage = c.GetImage(i.baseImageImageName).GetLastNonEmptyStage()
		i.baseImage = i.stageAsBaseImage.GetStageImage()
	} else {
		i.baseImageType = ImageFromRegistryAsBaseImage
		i.baseImage = c.GetOrCreateStageImage(nil, i.baseImageName)
	}
}

func (i *Image) GetBaseImage() *stage.StageImage {
	return i.baseImage
}

func (i *Image) FetchBaseImage(ctx context.Context, c *Conveyor) error {
	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		containerBackend := c.ContainerBackend

		if info, err := containerBackend.GetImageInfo(ctx, i.baseImage.Image.Name(), container_backend.GetImageInfoOpts{}); err != nil {
			return fmt.Errorf("unable to inspect local image %s: %w", i.baseImage.Image.Name(), err)
		} else if info != nil {
			// TODO: do not use container_backend.LegacyStageImage for base image
			i.baseImage.Image.SetStageDescription(&image.StageDescription{
				StageID: nil, // this is not a stage actually, TODO
				Info:    info,
			})

			baseImageRepoId, err := i.getFromBaseImageIdFromRegistry(ctx, c, i.baseImage.Image.Name())
			if baseImageRepoId == info.ID || err != nil {
				if err != nil {
					logboek.Context(ctx).Warn().LogF("WARNING: cannot get base image id (%s): %s\n", i.baseImage.Image.Name(), err)
					logboek.Context(ctx).Warn().LogF("WARNING: using existing image %s without pull\n", i.baseImage.Image.Name())
					logboek.Context(ctx).Warn().LogOptionalLn()
				}

				return nil
			}
		}

		if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", i.baseImage.Image.Name()).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				return c.ContainerBackend.PullImageFromRegistry(ctx, i.baseImage.Image)
			}); err != nil {
			return err
		}

		info, err := containerBackend.GetImageInfo(ctx, i.baseImage.Image.Name(), container_backend.GetImageInfoOpts{})
		if err != nil {
			return fmt.Errorf("unable to inspect local image %s: %w", i.baseImage.Image.Name(), err)
		}

		if info == nil {
			return fmt.Errorf("unable to inspect local image %s after successful pull: image is not exists", i.baseImage.Image.Name())
		}

		i.baseImage.Image.SetStageDescription(&image.StageDescription{
			StageID: nil, // this is not a stage actually, TODO
			Info:    info,
		})
	case StageAsBaseImage:
		// TODO: check no bug introduced
		// if err := c.ContainerBackend.RefreshImageObject(ctx, &container_backend.Image{Image: i.baseImage}); err != nil {
		//	return err
		// }
		if err := c.StorageManager.FetchStage(ctx, c.ContainerBackend, i.stageAsBaseImage); err != nil {
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

	switch {
	case i.baseImageRepoId != "":
		return i.baseImageRepoId, nil
	case c.IsBaseImagesRepoIdsCacheExist(baseImageName):
		i.baseImageRepoId = c.GetBaseImagesRepoIdsCache(baseImageName)
		return i.baseImageRepoId, nil
	case c.IsBaseImagesRepoErrCacheExist(baseImageName):
		return "", c.GetBaseImagesRepoErrCache(baseImageName)
	}

	var fetchedBaseRepoImage *image.Info
	processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", baseImageName)
	if err := logboek.Context(ctx).Info().LogProcessInline(processMsg).DoError(func() error {
		var fetchImageIdErr error
		fetchedBaseRepoImage, fetchImageIdErr = docker_registry.API().GetRepoImage(ctx, baseImageName)
		if fetchImageIdErr != nil {
			c.SetBaseImagesRepoErrCache(baseImageName, fetchImageIdErr)
			return fmt.Errorf("can not get base image id from registry (%s): %w", baseImageName, fetchImageIdErr)
		}

		return nil
	}); err != nil {
		return "", err
	}

	i.baseImageRepoId = fetchedBaseRepoImage.ID
	c.SetBaseImagesRepoIdsCache(baseImageName, i.baseImageRepoId)

	return i.baseImageRepoId, nil
}
