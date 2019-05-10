package build

import (
	"fmt"
	"strings"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/util"
)

const (
	BuildCacheVersion = "1"

	LocalImageStageImageNameFormat = "werf-stages-storage/%s"
	LocalImageStageImageFormat     = "werf-stages-storage/%s:%s"
)

func NewSignaturesPhase(lockImages bool) *SignaturesPhase {
	return &SignaturesPhase{LockImages: lockImages}
}

type SignaturesPhase struct {
	LockImages bool
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

		stageDependencies, err := s.GetDependencies(c, prevImage)
		if err != nil {
			return err
		}

		checksumArgs := []string{stageDependencies, BuildCacheVersion}

		if prevStage != nil {
			checksumArgs = append(checksumArgs, prevStage.GetSignature())
		}

		stageSig := util.Sha256Hash(checksumArgs...)

		s.SetSignature(stageSig)

		logboek.LogInfoF("%s:%s %s\n", s.Name(), strings.Repeat(" ", maxStageNameLength-len(s.Name())), stageSig)

		imageName := fmt.Sprintf(LocalImageStageImageFormat, c.projectName(), stageSig)

		i := c.GetOrCreateImage(prevImage, imageName)
		s.SetImage(i)

		if err = i.SyncDockerState(); err != nil {
			return fmt.Errorf("error synchronizing docker state of stage %s: %s", s.Name(), err)
		}

		if p.LockImages {
			imageLockName := imagePkg.ImageLockName(i.Name())
			if err := c.AcquireGlobalLock(imageLockName, lock.LockOptions{}); err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}

			if err = i.SyncDockerState(); err != nil {
				return fmt.Errorf("error synchronizing docker state of stage %s: %s", s.Name(), err)
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
