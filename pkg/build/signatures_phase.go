package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

const (
	BuildCacheVersion = "33"
)

func NewSignaturesPhase() *SignaturesPhase {
	return &SignaturesPhase{}
}

type SignaturesPhase struct{}

func (p *SignaturesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("SignaturesPhase.Run\n")
	}

	for _, dimg := range c.DimgsInOrder {
		if debug() {
			fmt.Printf("  dimg: '%s'\n", dimg.GetName())
		}

		var prevStage stage.Interface

		dimg.SetupBaseImage(c)

		var prevBuiltImage image.Image
		prevImage := dimg.GetBaseImage()
		err := prevImage.SyncDockerState()
		if err != nil {
			return err
		}

		var newStagesList []stage.Interface

		for _, s := range dimg.GetStages() {
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

			imageName := fmt.Sprintf("dimgstage-%s:%s", c.ProjectName, stageSig)
			i := c.GetOrCreateImage(prevImage, imageName)
			s.SetImage(i)

			err = i.SyncDockerState()
			if err != nil {
				return fmt.Errorf("error synchronizing docker state of stage %s: %s", s.Name(), err)
			}

			newStagesList = append(newStagesList, s)

			prevStage = s
			prevImage = i
		}

		dimg.SetStages(newStagesList)
	}

	return nil
}
