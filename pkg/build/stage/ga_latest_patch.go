package stage

import (
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func NewGALatestPatchStage(baseStageOptions *NewBaseStageOptions) *GALatestPatchStage {
	s := &GALatestPatchStage{}
	s.GAPatchStage = newGAPatchStage(baseStageOptions)
	return s
}

type GALatestPatchStage struct {
	*GAPatchStage
}

func (s *GALatestPatchStage) Name() StageName {
	return GALatestPatch
}

func (s *GALatestPatchStage) IsEmpty(_ Conveyor, prevBuiltImage image.Image) (bool, error) {
	if s.willLatestCommitBeBuiltOnGAArchiveStage(prevBuiltImage) {
		return true, nil
	}

	isEmpty := true
	for _, ga := range s.gitArtifacts {
		if empty, err := ga.IsPatchEmpty(prevBuiltImage); err != nil {
			return false, err
		} else if !empty {
			isEmpty = false
			break
		}
	}

	return isEmpty, nil
}

func (s *GALatestPatchStage) GetDependencies(_ Conveyor, prevImage image.Image) (string, error) {
	var args []string

	for _, ga := range s.gitArtifacts {
		commit, err := ga.LatestCommit()
		if err != nil {
			return "", err
		}

		args = append(args, commit)
	}

	return util.Sha256Hash(args...), nil
}
