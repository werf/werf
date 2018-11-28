package build

import (
	"reflect"

	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	c.DimgsInOrder = generateDimgsInOrder(c.Dappfile, c)
	return nil
}

func generateDimgsInOrder(dappfile []*config.Dimg, c *Conveyor) []*stage.Dimg {
	var dimgs []*stage.Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(dappfile) {
		dimg := &stage.Dimg{}
		dimg.SetStages(generateStages(dimgConfig, c))
		dimgs = append(dimgs, dimg)
	}

	return dimgs
}

func getDimgConfigsInOrder(dappfile []*config.Dimg) []interface{} {
	var dimgConfigs []interface{}
	for _, dimg := range dappfile {
		relatedDimgs := dimg.RelatedDimgs()
		for i := len(relatedDimgs) - 1; i > 0; i-- {
			if isNotInArr(dimgConfigs, relatedDimgs[i]) {
				dimgConfigs = append(dimgConfigs, relatedDimgs[i])
			}
		}
	}

	return dimgConfigs
}

func isNotInArr(arr []interface{}, obj interface{}) bool {
	for _, elm := range arr {
		if reflect.DeepEqual(elm, obj) {
			return false
		}
	}

	return true
}

func generateStages(dimgConfig interface{}, c *Conveyor) []stage.Interface {
	var stages []stage.Interface

	var dimgBase *config.DimgBase
	var dimgArtifact bool
	switch dimgConfig.(type) {
	case config.Dimg:
		dimgBase = dimgConfig.(*config.Dimg).DimgBase
		dimgArtifact = false
	case config.DimgArtifact:
		dimgBase = dimgConfig.(*config.DimgArtifact).DimgBase
		dimgArtifact = true
	}

	stages = appendIfExist(stages, stage.GenerateFromStage(dimgBase))

	ansibleBuilderExtra := &builder.Extra{
		TmpPath:           c.TmpDir,
		ContainerDappPath: c.ContainerDappPath,
	}

	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(dimgConfig, ansibleBuilderExtra))
	stages = appendIfExist(stages, stage.GenerateInstallStage(dimgConfig, ansibleBuilderExtra))
	stages = appendIfExist(stages, stage.GenerateBeforeSetupStage(dimgConfig, ansibleBuilderExtra))
	stages = appendIfExist(stages, stage.GenerateSetupStage(dimgConfig, ansibleBuilderExtra))

	if dimgArtifact {
		stages = appendIfExist(stages, stage.GenerateBuildArtifactStage(dimgConfig, ansibleBuilderExtra))
	} else {
		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(dimgConfig.(*config.Dimg)))
	}

	return stages
}

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if stage != nil {
		return append(stages, stage)
	}

	return stages
}
