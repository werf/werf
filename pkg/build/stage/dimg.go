package stage

import (
	"reflect"

	"github.com/flant/dapp/pkg/config"
)

type Dimg struct {
	stages []Interface
}

func (d *Dimg) SetStages(stages []Interface) {
	d.stages = stages
}

func (d *Dimg) GetStages() []Interface {
	return d.stages
}

func (d *Dimg) GetStage(name string) Interface {
	for _, stage := range d.stages {
		if stage.Name() == name {
			return stage
		}
	}

	return nil
}

func (d *Dimg) LatestStage() Interface {
	return d.stages[len(d.stages)-1]
}

func GenerateDimgsInOrder(dappfile []*config.Dimg) []*Dimg {
	var dimgs []*Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(dappfile) {
		stages := generateStages(dimgConfig)
		dimgs = append(dimgs, &Dimg{stages: stages})
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

func generateStages(dimgConfig interface{}) []Interface {
	var stages []Interface

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

	stages = appendIfExist(stages, generateFromStage(dimgBase))
	stages = appendIfExist(stages, generateBeforeInstallStage(dimgConfig))
	stages = appendIfExist(stages, generateInstallStage(dimgConfig))
	stages = appendIfExist(stages, generateBeforeSetupStage(dimgConfig))
	stages = appendIfExist(stages, generateSetupStage(dimgConfig))

	if dimgArtifact {
		stages = appendIfExist(stages, generateBuildArtifactStage(dimgConfig))
	} else {
		stages = appendIfExist(stages, generateDockerInstructionsStage(dimgConfig.(*config.Dimg)))
	}

	return stages
}

func appendIfExist(stages []Interface, stage Interface) []Interface {
	if stage != nil {
		return append(stages, stage)
	}

	return stages
}
