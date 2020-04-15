package stage

import "github.com/flant/werf/pkg/build/import_server"

type Conveyor interface {
	GetImageStageContentSignature(imageName, stageName string) string
	GetImageContentSignature(imageName string) string

	GetImageNameForLastImageStage(imageName string) string
	GetImageIDForLastImageStage(imageName string) string

	GetImageNameForImageStage(imageName, stageName string) string
	GetImageIDForImageStage(imageName, stageName string) string

	GetImportServer(imageName, stageName string) (import_server.ImportServer, error)
}
