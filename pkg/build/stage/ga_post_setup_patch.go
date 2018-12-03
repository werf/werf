package stage

import (
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

const patchSizeStep = 1024 * 1024

func NewGAPostSetupPatchStage(baseStageOptions *NewBaseStageOptions) *GAPostSetupPatchStage {
	s := &GAPostSetupPatchStage{}
	s.GAPatchStage = newGAPatchStage(baseStageOptions)
	return s
}

type GAPostSetupPatchStage struct {
	*GAPatchStage
}

func (s *GAPostSetupPatchStage) Name() StageName {
	return GAPostSetupPatch
}

func (s *GAPostSetupPatchStage) GetDependencies(_ Conveyor, prevImage image.Image) (string, error) {
	var size int64
	for _, ga := range s.gitArtifacts {
		commit, ok := prevImage.Labels()[ga.GetParamshash()]
		if ok {
			exist, err := ga.GitRepo().IsCommitExists(commit)
			if err != nil {
				return "", err
			}

			if exist {
				patchSize, err := ga.PatchSize(commit)
				if err != nil {
					return "", err
				}

				size += patchSize
			}
		}
	}

	return util.Sha256Hash(string(size / patchSizeStep)), nil
}
