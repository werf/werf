package stage

func NewGAArtifactPatchStage(baseStageOptions *NewBaseStageOptions) *GAPostInstallPatchStage {
	s := &GAPostInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage(baseStageOptions)
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
