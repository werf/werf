package stage

type Conveyor interface {
	GetDimgSignature(dimgName string) string
	GetDimgImageName(dimgName string) string
	SetBuildingGitStage(dimgName string, stageName StageName)
	GetBuildingGitStage(dimgName string) StageName
}
