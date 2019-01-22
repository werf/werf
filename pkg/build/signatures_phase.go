package build

import (
	"fmt"

	"github.com/flant/werf/pkg/build/stage"
	imagePkg "github.com/flant/werf/pkg/image"
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

func (p *SignaturesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("SignaturesPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if debug() {
			fmt.Printf("  image: '%s'\n", image.GetName())
		}

		var prevStage stage.Interface

		image.SetupBaseImage(c)

		var prevBuiltImage imagePkg.ImageInterface
		prevImage := image.GetBaseImage()
		err := prevImage.SyncDockerState()
		if err != nil {
			return err
		}

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
				continue
			}

			if debug() {
				fmt.Printf("    %s\n", s.Name())
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
			i := c.GetOrCreateImage(prevImage, imageName)
			s.SetImage(i)

			if err = i.SyncDockerState(); err != nil {
				return fmt.Errorf("error synchronizing docker state of stage %s: %s", s.Name(), err)
			}

			if err = s.AfterImageSyncDockerStateHook(c); err != nil {
				return err
			}

			if image.GetName() == "" {
				fmt.Printf("# Calculated signature %s for image %s\n", stageSig, fmt.Sprintf("stage/%s", s.Name()))
			} else {
				fmt.Printf("# Calculated signature %s for image/%s %s\n", stageSig, image.GetName(), fmt.Sprintf("stage/%s", s.Name()))
			}

			newStagesList = append(newStagesList, s)

			prevStage = s
			prevImage = i
		}

		image.SetStages(newStagesList)
	}

	return nil
}
