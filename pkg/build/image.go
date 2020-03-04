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

	stages            []stage.Interface
	lastNonEmptyStage stage.Interface
	stagesSignature   string
	baseImage         *image.StageImage
	isArtifact        bool
	isDockerfileImage bool
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

func (i *Image) SetStagesSignature(sig string) {
	i.stagesSignature = sig
}

func (i *Image) GetStagesSignature() string {
	return i.stagesSignature
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
	baseImageName := i.baseImageName
	if i.baseImageImageName != "" {
		baseImageName = c.GetImage(i.baseImageImageName).GetLastNonEmptyStage().GetImage().Name()
	}

	i.baseImage = c.GetOrCreateStageImage(nil, baseImageName)
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
				logboek.LogWarnF("WARNING: cannot get base image id (%s): %s\n", i.baseImage.Name(), err)
				logboek.LogWarnF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
				logboek.Warn.LogOptionalLn()
			}

			return nil
		}
	}

	logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
	return logboek.Default.LogProcess("Pulling base image", logProcessOptions, func() error {
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
	if err := logboek.Info.LogProcessInline(processMsg, logboek.LevelLogProcessInlineOptions{}, func() error {
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
