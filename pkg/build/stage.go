package build

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
	return stage.LayerCommitMap[gitArtifact.Paramshash], nil
}
