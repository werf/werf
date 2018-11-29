package stage

import "github.com/flant/dapp/pkg/util"

func NewGALatestPatchStage() *GALatestPatchStage {
	s := &GALatestPatchStage{}
	s.GAStage = newGAStage()

	return s
}

type GALatestPatchStage struct {
	*GAStage
}

func (s *GALatestPatchStage) Name() StageName {
	return GALatestPatch
}

func (s *GALatestPatchStage) GetDependencies(_ Conveyor, prevImage Image) (string, error) {
	var args []string

	for _, ga := range s.gitArtifacts {
		isEmpty, err := ga.IsPatchEmpty(prevImage)
		if err != nil {
			return "", err
		}

		if !isEmpty {
			commit, err := ga.LatestCommit()
			if err != nil {
				return "", err
			}

			args = append(args, commit)
		}
	}

	return util.Sha256Hash(args...), nil
}
