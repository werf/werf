package build

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/build/stage"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/util"
)

const (
	BuildCacheVersion = "33"

	LocalImageStageImageNameFormat = "image-stage-%s"
	LocalImageStageImageFormat     = "image-stage-%s:%s"
)

func NewSignaturesPhase() *SignaturesPhase {
	return &SignaturesPhase{}
}

type SignaturesPhase struct{}

func (p *SignaturesPhase) Run(c *Conveyor) (err error) {
	for ind, image := range c.imagesInOrder {
		isLastImage := ind == len(c.imagesInOrder)-1

		err = logger.LogServiceProcess(fmt.Sprintf("Calculate %s signatures", image.LogName()), "", func() error {
			return p.calculateImageSignatures(c, image)
		})

		if err != nil {
			return err
		}

		if !isLastImage {
			fmt.Println()
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
			logger.LogInfoF("%s:%s <empty>\n", s.Name(), strings.Repeat(" ", maxStageNameLength-len(s.Name())))
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

		imageName := fmt.Sprintf(LocalImageStageImageFormat, c.projectName(), stageSig)

		logger.LogInfoF("%s:%s %s\n", s.Name(), strings.Repeat(" ", maxStageNameLength-len(s.Name())), imageName)

		i := c.GetOrCreateImage(prevImage, imageName)
		s.SetImage(i)

		if err = i.SyncDockerState(); err != nil {
			return fmt.Errorf("error synchronizing docker state of stage %s: %s", s.Name(), err)
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
		fmt.Println()
		logger.LogInfoF("Git files are actual on the %s stage\n", stageName)
	}

	image.SetStages(newStagesList)

	return nil
}
