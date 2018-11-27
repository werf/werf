package build

import "fmt"

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

	/*
		for c.GetDimgsInOrder {
			for dimg.GetStages() {
				stageSig = stage.GetDependecies + prevStageSignature + BuildCacheVersion + getStage(stage.RelatedStageName).GetContext
				stage.SetSignature(stageSig)
				prevStageSignature = stageSig

				create image by stageSig
				c.GetOrCreateImage(fromImage *Stage, name string)

			}
		}
	*/

	return nil
}
