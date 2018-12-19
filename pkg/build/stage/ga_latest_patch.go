package stage

import (
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func NewGALatestPatchStage(gaPatchStageOptions *NewGaPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GALatestPatchStage {
	s := &GALatestPatchStage{}
	s.GAPatchStage = newGAPatchStage(GALatestPatch, gaPatchStageOptions, baseStageOptions)
	return s
}

type GALatestPatchStage struct {
	*GAPatchStage
}

func (s *GALatestPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error) {
	if empty, err := s.GAPatchStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	isEmpty := true
	for _, ga := range s.gitArtifacts {
		commit := ga.GetGACommitFromImageLabels(prevBuiltImage)
		if exist, err := ga.GitRepo().IsCommitExists(commit); err != nil {
			return false, err
		} else if !exist {
			return true, nil
		}

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
