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

	for _, dimg := range c.GetDimgsInOrder() {
		var prevStage stage.Interface
		var prevImage *image.Stage

		newStagesList := []stage.Interface{}

		for _, stage := range dimg.GetStages() {
			isEmpty, err := stage.IsEmpty(c, prevImage)
			if err != nil {
				return fmt.Errorf("error checking stage %s emptyness: %s", stage.Name(), err)
			}
			if isEmpty {
				continue
			}

			stageDependencies, err := stage.GetDependencies(c, prevImage)
			if err != nil {
				return err
			}

			checksumArgs := []string{stageDependencies, BuildCacheVersion}

			if prevStage != nil {
				checksumArgs = append(checksumArgs, prevStage.GetSignature())
			}

			relatedStage := dimg.GetStage(stage.GetRelatedStageName())
			// related stage may be empty
			if relatedStage != nil {
				relatedStageContext, err := relatedStage.GetContext(c)
				if err != nil {
					return err
				}

				checksumArgs = append(checksumArgs, relatedStageContext)
			}

			stageSig := util.Sha256Hash(checksumArgs...)

			stage.SetSignature(stageSig)

			imageName := fmt.Sprintf("dimgstage-%s:%s", c.GetProjectName(), stageSig)
			image := c.GetOrCreateImage(prevImage, imageName)
			stage.SetImage(image)

			err = image.ReadDockerState()
			if err != nil {
				return fmt.Errorf("error reading docker state of stage %s: %s", stage.Name(), err)
			}

			newStagesList = append(newStagesList, stage)

			prevStage = stage
			prevImage = image
		}

		dimg.SetStages(newStagesList)
	}

	return nil
}
