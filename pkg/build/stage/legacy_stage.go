package stage

import (
	"fmt"

	"github.com/flant/dapp/pkg/image"
)

type LegacyStage interface {
	GetPrevStage() LegacyStage
	GetImage() image.Image
	LayerCommit(gitArtifact *GitArtifact) (string, error)
}

type StubStage struct {
	LayerCommitMap map[string]string
	PrevStage      *StubStage
	Image          *StubImage
}

func (stage *StubStage) GetPrevStage() LegacyStage {
	return stage.PrevStage
}

func (stage *StubStage) GetImage() image.Image {
	return stage.Image
}

func (stage *StubStage) LayerCommit(gitArtifact *GitArtifact) (string, error) {
	if commit, hasKey := stage.LayerCommitMap[gitArtifact.ImageGACommitLabel()]; hasKey {
		return commit, nil
	}

	panic(fmt.Errorf("assertion failed: StubStage layer commit should be present for git `%s`", gitArtifact.GetParamshash()))
}
