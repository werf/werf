package build

import "fmt"

type Stage interface {
	GetPrevStage() Stage
	GetImage() Image
	LayerCommit(gitArtifact *GitArtifact) (string, error)
}

type StubStage struct {
	LayerCommitMap map[string]string
	PrevStage      *StubStage
	Image          *StubImage
}

func (stage *StubStage) GetPrevStage() Stage {
	return stage.PrevStage
}

func (stage *StubStage) GetImage() Image {
	return stage.Image
}

func (stage *StubStage) LayerCommit(gitArtifact *GitArtifact) (string, error) {
	if commit, hasKey := stage.LayerCommitMap[gitArtifact.Paramshash]; hasKey {
		return commit, nil
	}

	panic(fmt.Errorf("assertion failed: StubStage layer commit should be present for git `%s`", gitArtifact.Paramshash))
}
