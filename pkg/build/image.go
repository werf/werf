package build

import (
	"fmt"

	"github.com/flant/werf/pkg/container_runtime"

	"github.com/fatih/color"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logging"
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
	contentSignature  string
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

func (i *Image) LogProcessStyle() *logboek.Style {
	return ImageLogProcessStyle(i.isArtifact)
}

func (i *Image) LogTagStyle() *logboek.Style {
	return ImageLogTagStyle(i.isArtifact)
}

func ImageLogProcessStyle(isArtifact bool) *logboek.Style {
	return imageDefaultStyle(isArtifact)
}

func ImageLogTagStyle(isArtifact bool) *logboek.Style {
	return imageDefaultStyle(isArtifact)
}

func imageDefaultStyle(isArtifact bool) *logboek.Style {
	var attributes []color.Attribute
	if isArtifact {
		attributes = []color.Attribute{color.FgCyan, color.Bold}
	} else {
		attributes = []color.Attribute{color.FgYellow, color.Bold}
	}

	return &logboek.Style{Attributes: attributes}
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

func (i *Image) SetContentSignature(sig string) {
	i.contentSignature = sig
}

func (i *Image) GetContentSignature() string {
	return i.contentSignature
}

func (i *Image) GetStage(name stage.StageName) stage.Interface {
	for _, s := range i.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
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

func (i *Image) CleanupBaseImage(c *Conveyor) error {
	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		// Do not cleanup pulled images from registry
		return nil
	case StageAsBaseImage:
		if shouldCleanup, err := c.StagesStorage.ShouldCleanupLocalImage(&container_runtime.DockerImage{Image: i.stageAsBaseImage.GetImage()}); err == nil && shouldCleanup {
			if err := logboek.Default.LogProcess(
				fmt.Sprintf("Cleaning up stage %q local image", i.stageAsBaseImage.LogDetailedName()),
				logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
				func() error {
					logboek.Info.LogF("Image name: %s\n", i.stageAsBaseImage.GetImage().Name())
					if err := c.StagesStorage.CleanupLocalImage(&container_runtime.DockerImage{Image: i.stageAsBaseImage.GetImage()}); err != nil {
						return fmt.Errorf("unable to cleanup stage %q local image %s for stages storage %s: %s", i.stageAsBaseImage.LogDetailedName(), i.stageAsBaseImage.GetImage().Name(), c.StagesStorage.String(), err)
					}
					return nil
				},
			); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		return nil
	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}
}

func (i *Image) FetchBaseImage(c *Conveyor) error {
	switch i.baseImageType {
	case ImageFromRegistryAsBaseImage:
		if err := c.ContainerRuntime.RefreshImageObject(&container_runtime.DockerImage{Image: i.baseImage}); err != nil {
			return err
		}

		if i.baseImage.IsExistsLocally() {
			baseImageRepoId, err := i.getFromBaseImageIdFromRegistry(c, i.baseImage.Name())
			if baseImageRepoId == i.baseImage.GetStagesStorageImageInfo().ID || err != nil {
				if err != nil {
					logboek.LogWarnF("WARNING: cannot get base image id (%s): %s\n", i.baseImage.Name(), err)
					logboek.LogWarnF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
					logboek.Warn.LogOptionalLn()
				}

				return nil
			}
		}

		logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
		return logboek.Default.LogProcess(fmt.Sprintf("Pulling base image %s", i.baseImage.Name()), logProcessOptions, func() error {
			return c.ContainerRuntime.PullImageFromRegistry(&container_runtime.DockerImage{Image: i.baseImage})
		})
	case StageAsBaseImage:
		if err := c.ContainerRuntime.RefreshImageObject(&container_runtime.DockerImage{Image: i.baseImage}); err != nil {
			return err
		}

		if shouldFetch, err := c.StagesStorage.ShouldFetchImage(&container_runtime.DockerImage{Image: i.baseImage}); err == nil && shouldFetch {
			return logboek.Default.LogProcess(
				fmt.Sprintf("Fetching base stage %q image from stages storage", i.stageAsBaseImage.LogDetailedName()),
				logboek.LevelLogProcessOptions{}, func() error {
					logboek.Info.LogF("Image name: %s\n", i.stageAsBaseImage.GetImage().Name())
					if err := c.StagesStorage.FetchImage(&container_runtime.DockerImage{Image: i.stageAsBaseImage.GetImage()}); err != nil {
						return fmt.Errorf("unable to fetch stage %q image %s from stages storage %s: %s", i.stageAsBaseImage.LogDetailedName(), i.stageAsBaseImage.GetImage().Name(), c.StagesStorage.String(), err)
					}
					return nil
				})
		}
	default:
		panic(fmt.Sprintf("unknown base image type %q", i.baseImageType))
	}

	return nil
}

func (i *Image) getFromBaseImageIdFromRegistry(c *Conveyor, baseImageName string) (string, error) {
	if i.baseImageRepoId != "" {
		return i.baseImageRepoId, nil
	} else if cachedBaseImageRepoId, exist := c.baseImagesRepoIdsCache[baseImageName]; exist {
		i.baseImageRepoId = cachedBaseImageRepoId
		return cachedBaseImageRepoId, nil
	} else if cachedBaseImagesRepoErr, exist := c.baseImagesRepoErrCache[baseImageName]; exist {
		return "", cachedBaseImagesRepoErr
	}

	var fetchedBaseRepoImage *image.Info
	processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", baseImageName)
	if err := logboek.Info.LogProcessInline(processMsg, logboek.LevelLogProcessInlineOptions{}, func() error {
		var fetchImageIdErr error
		fetchedBaseRepoImage, fetchImageIdErr = docker_registry.API().GetRepoImage(baseImageName)
		if fetchImageIdErr != nil {
			c.baseImagesRepoErrCache[baseImageName] = fetchImageIdErr
			return fmt.Errorf("can not get base image id from registry (%s): %s", baseImageName, fetchImageIdErr)
		}

		return nil
	}); err != nil {
		return "", err
	}

	i.baseImageRepoId = fetchedBaseRepoImage.ID
	c.baseImagesRepoIdsCache[baseImageName] = i.baseImageRepoId

	return i.baseImageRepoId, nil
}
