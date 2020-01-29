package build

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

func NewSignaturesPhase(lockStages bool) *SignaturesPhase {
	return &SignaturesPhase{LockStages: lockStages}
}

type SignaturesPhase struct {
	LockStages bool
}

func (p *SignaturesPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Calculating signatures", logProcessOptions, func() error {
		return logboek.WithoutIndent(func() error { return p.run(c) })
	})
}

func (p *SignaturesPhase) run(c *Conveyor) error {
	for _, image := range c.imagesInOrder {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.calculateImageSignatures(c, image)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *SignaturesPhase) calculateImageSignatures(c *Conveyor, image *Image) error {
	var prevStage stage.Interface

	image.SetupBaseImage(c)

	var prevBuiltImage imagePkg.ImageInterface
	prevImage := image.GetBaseImage()
	err := prevImage.SyncDockerState()
	if err != nil {
		return err
	}

	maxStageNameLength := 22

	var newStagesList []stage.Interface

	for _, s := range image.GetStages() {
		if prevImage.IsExists() {
			prevBuiltImage = prevImage
		}

		isEmpty, err := s.IsEmpty(c, prevBuiltImage)
		if err != nil {
			return fmt.Errorf("error checking stage %s is empty: %s", s.Name(), err)
		}
		if isEmpty {
			logboek.LogInfoF("%s:%s <empty>\n", s.Name(), strings.Repeat(" ", maxStageNameLength-len(s.Name())))
			continue
		}

		stageDependencies, err := s.GetDependencies(c, prevImage, prevBuiltImage)
		if err != nil {
			return err
		}

		checksumArgs := []string{stageDependencies, imagePkg.BuildCacheVersion}
		if prevStage != nil {
			prevStageDependencies, err := prevStage.GetNextStageDependencies(c)
			if err != nil {
				return fmt.Errorf("unable to get prev stage %s dependencies for the stage %s: %s", prevStage.Name(), s.Name(), err)
			}

			checksumArgs = append(checksumArgs, prevStage.GetSignature(), prevStageDependencies)
		}
		stageSig := util.Sha256Hash(checksumArgs...)
		s.SetSignature(stageSig)

		logboek.LogInfoF("%s:%s %s\n", s.Name(), strings.Repeat(" ", maxStageNameLength-len(s.Name())), stageSig)

		imagesDescs, err := c.StagesStorage.GetImagesBySignature(c.projectName(), stageSig)
		if err != nil {
			return fmt.Errorf("unable to get images from stages storage %s by signature %s: %s", c.StagesStorage.String(), stageSig)
		}

		var imageExists bool
		var i *imagePkg.StageImage

		if len(imagesDescs) > 0 {
			imgInfo, err := s.SelectCacheImage(imagesDescs)
			if err != nil {
				return err
			}

			if imgInfo != nil {
				fmt.Printf("-- SelectCacheImage => %v\n", imgInfo)
				imageExists = true

				i = c.GetOrCreateImage(prevImage, imgInfo.ImageName)
				s.SetImage(i)

				if err := c.StagesStorage.SyncStageImage(i); err != nil {
					return fmt.Errorf("unable to fetch image %s from stages storage %s: %s", imgInfo.ImageName, c.StagesStorage.String(), err)
				}
			}
		}

		if !imageExists {
			imageName := fmt.Sprintf(imagePkg.LocalImageStageImageFormat, c.projectName(), stageSig, util.UUIDToShortString(uuid.New()))

			i = c.GetOrCreateImage(prevImage, imageName)
			s.SetImage(i)

			if p.LockStages {
				if err := c.StagesStorageLockManager.LockStage(c.projectName(), stageSig); err != nil {
					return fmt.Errorf("failed to lock %s: %s", stageSig, err)
				}
			}
		}

		if err = s.AfterImageSyncDockerStateHook(c); err != nil {
			return err
		}

		newStagesList = append(newStagesList, s)

		prevStage = s
		prevImage = i
	}

	stageName := c.GetBuildingGitStage(image.name)
	if stageName != "" {
		logboek.LogLn()
		logboek.LogInfoF("Git files will be actualized on stage %s\n", stageName)
	}

	image.SetStages(newStagesList)

	return nil
}
