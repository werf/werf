package stage

type Conveyor interface {
	GetImageStagesSignature(imageName string) string
	GetImageLastStageImageName(imageName string) string
	GetImageLastStageImageID(imageName string) string
	SetBuildingGitStage(imageName string, stageName StageName)
	GetBuildingGitStage(imageName string) StageName
}
