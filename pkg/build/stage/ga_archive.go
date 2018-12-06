package stage

import (
	"sort"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

const GAArchiveResetCommitRegex = "(\\[dapp reset\\])|(\\[reset dapp\\])"

func NewGAArchiveStage(baseStageOptions *NewBaseStageOptions) *GAArchiveStage {
	s := &GAArchiveStage{}
	s.GAStage = newGAStage(GAArchive, baseStageOptions)
	return s
}

type GAArchiveStage struct {
	*GAStage
}

func (s *GAArchiveStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	var args []string
	for _, ga := range s.gitArtifacts {
		args = append(args, ga.GetParamshash())

		commit, err := ga.GitRepo().FindCommitIdByMessage(GAArchiveResetCommitRegex)
		if err != nil {
			return "", err
		}

		args = append(args, commit)
	}

	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *GAArchiveStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyArchiveCommand(image); err != nil {
			return err
		}
	}

	return nil
}
