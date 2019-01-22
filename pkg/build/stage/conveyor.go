package stage

type Conveyor interface {
	GetImageLatestStageSignature(imageName string) string
	GetImageLatestStageImageName(imageName string) string
	SetBuildingGitStage(imageName string, stageName StageName)
	GetBuildingGitStage(imageName string) StageName
}
