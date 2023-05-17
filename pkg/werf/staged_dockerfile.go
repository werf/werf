package werf

type StagedDockerfileVersion string

const (
	StagedDockerfileV1 StagedDockerfileVersion = "StagedDockerfileV1"
	StagedDockerfileV2 StagedDockerfileVersion = "StagedDockerfileV2"
)

var stagedDockerfileVersion StagedDockerfileVersion

func GetStagedDockerfileVersion() StagedDockerfileVersion {
	return stagedDockerfileVersion
}
