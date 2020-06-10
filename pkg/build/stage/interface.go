package stage

import (
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	Name() StageName
	LogDetailedName() string

	IsEmpty(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error)

	GetDependencies(c Conveyor, prevImage container_runtime.ImageInterface, prevBuiltImage container_runtime.ImageInterface) (string, error)
	GetNextStageDependencies(c Conveyor) (string, error)

	PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error

	PreRunHook(Conveyor) error

	SetSignature(signature string)
	GetSignature() string

	SetContentSignature(contentSignature string)
	GetContentSignature() string

	SetImage(container_runtime.ImageInterface)
	GetImage() container_runtime.ImageInterface

	SetGitMappings([]*GitMapping)
	GetGitMappings() []*GitMapping

	SelectSuitableStage(c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error)
}
