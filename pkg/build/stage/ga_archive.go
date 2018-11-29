package stage

import (
	"github.com/flant/dapp/pkg/util"
)

const GAArchiveResetCommitRegex = "(\\[dapp reset\\])|(\\[reset dapp\\])"

func NewGAArchiveStage() *GAArchiveStage {
	s := &GAArchiveStage{}
	s.GAStage = newGAStage()

	return s
}

type GAArchiveStage struct {
	*GAStage
}

func (s *GAArchiveStage) Name() StageName {
	return GAArchive
}

func (s *GAArchiveStage) GetDependencies(_ Conveyor, _ Image) string {
	var args []string
	for _, ga := range s.gitArtifacts {
		args = append(args, ga.GetParamshash())

		commit, err := ga.GitRepo().FindCommitIdByMessage(GAArchiveResetCommitRegex)
		if err != nil {
			panic(err)
		}

		args = append(args, commit)
	}

	return util.Sha256Hash(args...)
}
