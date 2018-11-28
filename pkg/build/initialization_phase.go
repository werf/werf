package build

import (
	"reflect"

	"github.com/flant/dapp/pkg/config"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	c.DimgsInOrder = generateDimgsInOrder(c)
	return nil
}

type Dimg struct {
	stages []Stage
}

func (d *Dimg) LatestStage() Stage {
	return d.stages[len(d.stages)]
}

func (d *Dimg) GetImage() Image {
	return d.LatestStage().GetImage()
}

func generateDimgsInOrder(c *Conveyor) []*Dimg {
	var dimgs []*Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(c) {
		stages := generateStages(dimgConfig)
		dimgs = append(dimgs, &Dimg{stages: stages})
	}

	return dimgs
}

func getDimgConfigsInOrder(c *Conveyor) []interface{} {
	var dimgConfigs []interface{}
	for _, dimg := range c.Dappfile {
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

func generateStages(c interface{}) []Stage {
	var stages []Stage

	var dimgBase *config.DimgBase
	switch c.(type) {
	case config.Dimg:
		dimgBase = c.(config.Dimg).DimgBase
	case config.DimgArtifact:
		dimgBase = c.(config.DimgArtifact).DimgBase
	}

	stages = append(stages, generateFromStage(dimgBase))
	//stages = append(stages, GenerateBeforeInstallStage(dimgBase))

	return nil
}

func generateFromStage(c *config.DimgBase) Stage {
	if c.From != "" {
		return FromStage{}
	} else if c.FromDimg != nil {
		return FromDimgStage{}
	} else if c.FromDimgArtifact != nil {
		return FromDimgArtifactStage{}
	} else {
		return nil
	}
}

//func GenerateBeforeInstallStage(c *config.DimgBase) Stage {
//	if c.Ansible != nil && len(c.Ansible.BeforeInstall) != 0 {
//		return AnsibleBeforeInstallStage{}
//	} else if c. != nil && len(c.Ansible.BeforeInstall) != 0 {
//
//	}
//}

type BaseStage struct{ Stage }

type FromStage struct{ BaseStage }
type FromDimgBaseStage struct {
	BaseStage

	FromDimg *Dimg
}

func (s *FromDimgBaseStage) BaseImage() Image {
	return s.FromDimg.GetImage()
}

type FromDimgStage struct {
	FromDimgBaseStage
}
type FromDimgArtifactStage struct {
	FromDimgBaseStage
}

type DockerStage struct{ BaseStage }

type UserStage struct{ BaseStage }

type BeforeInstallStage struct{ UserStage }
type InstallStage struct{ UserStage }
type BeforeSetupStage struct{ UserStage }
type SetupStage struct{ UserStage }
type BuildArtifactStage struct{ UserStage }

type AnsibleUserStage struct{ BaseStage }
type AnsibleBeforeInstallStage struct {
	BeforeInstallStage
	AnsibleUserStage
}
type AnsibleInstallStage struct {
	InstallStage
	AnsibleUserStage
}
type AnsibleBeforeSetupStage struct {
	BeforeSetupStage
	AnsibleUserStage
}
type AnsibleSetupStage struct {
	SetupStage
	AnsibleUserStage
}
type AnsibleBuildArtifactStage struct {
	BuildArtifactStage
	AnsibleUserStage
}

type ShellStage struct{ BaseStage }
type ShellBeforeInstallStage struct {
	BeforeInstallStage
	ShellStage
}
type ShellInstallStage struct {
	InstallStage
	ShellStage
}
type ShellBeforeSetupStage struct {
	BeforeSetupStage
	ShellStage
}
type ShellSetupStage struct {
	SetupStage
	ShellStage
}
type ShellBuildArtifactStage struct {
	BuildArtifactStage
	ShellStage
}

type GAStage struct{ BaseStage }

type GAArchiveStage struct{ GAStage }
type GABeforeInstallPatchStage struct{ GAStage }
type GAAfterInstallPatchStage struct{ GAStage }
type GABeforeSetupPatchStage struct{ GAStage }
type GASetupPatchStage struct{ GAStage }
type GALatestPatchStage struct{ GAStage }
type GAArtifactPatchStage struct{ GAStage }
