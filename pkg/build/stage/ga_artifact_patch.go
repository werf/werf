package stage

func NewGAArtifactPatchStage() *GAPostInstallPatchStage {
	s := &GAPostInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage()

	return s
}

type GAArtifactPatchStage struct {
	*GARelatedStage
}

func (s *GAArtifactPatchStage) Name() StageName {
	return GAArtifactPatch
}

func (s *GAArtifactPatchStage) GetRelatedStageName() StageName {
	return BuildArtifact
}
