package build

import (
	"fmt"
	"strings"

	"github.com/flant/dapp/pkg/build/stage"
)

func NewPrepareImagesPhase() *PrepareImagesPhase {
	return &PrepareImagesPhase{}
}

type PrepareImagesPhase struct{}

func (p *PrepareImagesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PrepareImagesPhase.Run\n")
	}

	for _, dimg := range c.GetDimgsInOrder() {
		var prevStage stage.Interface

		for _, stage := range dimg.GetStages() {
			err := p.AddMounts(dimg, stage, prevStage)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *PrepareImagesPhase) AddMounts(dimg *Dimg, stage stage.Interface, prevStage stage.Interface) error {
	mountpointsByType := map[string][]string{}

	for _, mountCfg := range dimg.GetConfig().Mount {
		mountpointsByType[mountCfg.Type] = append(mountpointsByType[mountCfg.Type], mountCfg.To)
	}

	labels := prevStage.GetImage().GetLabels()
	for _, labelMountType := range []struct{ Label, MountType string }{
		struct{ Label, MountType string }{"dapp-mount-tmp-dir", "tmp_dir"},
		struct{ Label, MountType string }{"dapp-mount-build-dir", "build_dir"},
	} {
		value, hasKey := labels[labelMountType.Label]
		if !hasKey {
			continue
		}

		mountpoints := strings.Split(value, ";")
		for _, mountpoint := range mountpoints {
			if mountpoint == "" {
				continue
			}

			mountpointsByType[labelMountType.MountType] = append(mountpointsByType[labelMountType.MountType], mountpoint)
		}
	}

	return nil
}
