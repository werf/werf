package stage

import "github.com/flant/werf/pkg/build/import_server"

type Conveyor interface {
	GetImageStagesSignature(imageName string) string
	GetImageLastStageImageName(imageName string) string
	GetImageLastStageImageID(imageName string) string
	SetBuildingGitStage(imageName string, stageName StageName)
	GetBuildingGitStage(imageName string) StageName
	GetImportServer(imageName string) (import_server.ImportServer, error)
}
