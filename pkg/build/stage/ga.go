package stage

import "github.com/flant/dapp/pkg/image"

func newGAStage(baseStageOptions *NewBaseStageOptions) *GAStage {
	s := &GAStage{}
	s.BaseStage = newBaseStage(baseStageOptions)
	return s
}

type GAStage struct {
	*BaseStage
}

func (s *GAStage) willLatestCommitBeBuiltOnPrevStage(prevBuiltImage image.Image) bool {
	if prevBuiltImage == nil {
		return true
	}

	for _, ga := range s.gitArtifacts {
		_, exist := prevBuiltImage.Labels()[ga.GetParamshash()]
		if !exist {
			return false
		}
	}

	return true
}
