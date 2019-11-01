package build

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logging"
)

type Image struct {
	name string

	baseImageName      string
	baseImageImageName string
	baseImageRepoId    string

	stages     []stage.Interface
	baseImage  *image.StageImage
	isArtifact bool
}

func (i *Image) LogName() string {
	return logging.ImageLogName(i.name, i.isArtifact)
}

func (i *Image) LogDetailedName() string {
	return logging.ImageLogProcessName(i.name, i.isArtifact)
}

func (i *Image) LogProcessColorizeFunc() func(...interface{}) string {
	return ImageLogProcessColorizeFunc(i.isArtifact)
}

func (i *Image) LogTagColorizeFunc() func(...interface{}) string {
	return ImageLogTagColorizeFunc(i.isArtifact)
}

func ImageLogProcessColorizeFunc(isArtifact bool) func(...interface{}) string {
	return imageDefaultColorizeFunc(isArtifact)
}

func ImageLogTagColorizeFunc(isArtifact bool) func(...interface{}) string {
	return imageDefaultColorizeFunc(isArtifact)
}

func imageDefaultColorizeFunc(isArtifact bool) func(...interface{}) string {
	var colorFormat []color.Attribute
	if isArtifact {
		colorFormat = []color.Attribute{color.FgCyan, color.Bold}
	} else {
		colorFormat = []color.Attribute{color.FgYellow, color.Bold}
	}

	return color.New(colorFormat...).Sprint
}

func (i *Image) SetStages(stages []stage.Interface) {
	i.stages = stages
}

func (i *Image) GetStages() []stage.Interface {
	return i.stages
}

func (i *Image) GetStage(name stage.StageName) stage.Interface {
	for _, s := range i.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
}

func (i *Image) LatestStage() stage.Interface {
	return i.stages[len(i.stages)-1]
}

func (i *Image) GetName() string {
	return i.name
}

func (i *Image) SetupBaseImage(c *Conveyor) {
	baseImageName := i.baseImageName
	if i.baseImageImageName != "" {
		baseImageName = c.GetImage(i.baseImageImageName).LatestStage().GetImage().Name()
	}

	i.baseImage = c.GetOrCreateImage(nil, baseImageName)
}

func (i *Image) GetBaseImage() *image.StageImage {
	return i.baseImage
}

func (i *Image) PrepareBaseImage(c *Conveyor) error {
	fromImage := i.stages[0].GetImage()

	if fromImage.IsExists() {
		return nil
	}

	if i.baseImageImageName != "" {
		return nil
	}

	if i.baseImage.IsExists() {
		baseImageRepoId, err := i.getFromBaseImageIdFromRegistry(c, i.baseImage.Name())
		if baseImageRepoId == i.baseImage.ID() || err != nil {
			if err != nil {
				logboek.LogLn()
				logboek.LogErrorF("WARNING: cannot get base image id (%s): %s\n", i.baseImage.Name(), err)
				logboek.LogErrorF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
			}

			return nil
		}

		logboek.LogOptionalLn()
	}

	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Pulling base image", logProcessOptions, func() error {
		if err := i.baseImage.Pull(); err != nil {
			return err
		}

		if err := i.baseImage.SyncDockerState(); err != nil {
			return err
		}

		return nil
	})
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

	var fetchedBaseImageRepoId string
	processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", baseImageName)
	if err := logboek.LogProcessInline(processMsg, logboek.LogProcessInlineOptions{}, func() error {
		var fetchImageIdErr error
		fetchedBaseImageRepoId, fetchImageIdErr = docker_registry.ImageId(baseImageName)
		if fetchImageIdErr != nil {
			c.baseImagesRepoErrCache[baseImageName] = fetchImageIdErr
			return fmt.Errorf("can not get base image id from registry (%s): %s", baseImageName, fetchImageIdErr)
		}

		return nil
	}); err != nil {
		return "", err
	}

	i.baseImageRepoId = fetchedBaseImageRepoId
	c.baseImagesRepoIdsCache[baseImageName] = fetchedBaseImageRepoId

	return i.baseImageRepoId, nil
}
