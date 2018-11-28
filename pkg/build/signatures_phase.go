package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/build/stage"
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
		var prevImage stage.Image

		for _, stage := range dimg.GetStages() {
			checksumArgs := []string{stage.GetDependencies(c), BuildCacheVersion}

			if prevStage != nil {
				checksumArgs = append(checksumArgs, prevStage.GetSignature())
			}

			relatedStage := dimg.GetStage(stage.GetRelatedStageName())
			// related stage may be empty
			if relatedStage != nil {
				checksumArgs = append(checksumArgs, relatedStage.GetContext(c))
			}

			stageSig := util.Sha256Hash(checksumArgs...)

			stage.SetSignature(stageSig)

			imageName := fmt.Sprintf("dimgstage-%s:%s", c.GetProjectName(), stageSig)
			image := c.GetOrCreateImage(prevImage, imageName)
			stage.SetImage(image)

			prevStage = stage
			prevImage = image
		}
	}

	return nil
}
