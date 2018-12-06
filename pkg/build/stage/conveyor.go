package stage

type Conveyor interface {
	GetDimgSignature(dimgName string) string
	GetDimgImageName(dimgName string) string
	SetBuildingGAStage(dimgName string, stageName StageName)
	GetBuildingGAStage(dimgName string) StageName
}
