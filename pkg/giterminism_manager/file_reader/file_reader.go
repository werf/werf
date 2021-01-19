package file_reader

import "github.com/werf/werf/pkg/git_repo"

type configType string

const (
	giterminismConfigErrorConfigType configType = "giterminism config"
	configErrorConfigType            configType = "werf config"
	configTemplateErrorConfigType    configType = "werf config template"
	configGoTemplateErrorConfigType  configType = "file"
	dockerfileErrorConfigType        configType = "dockerfile"
	dockerignoreErrorConfigType      configType = "dockerignore file"
	chartFileErrorConfigType         configType = "chart file"
	chartDirectoryErrorConfigType    configType = "chart directory"
)

type FileReader struct {
	sharedOptions     sharedOptions
	giterminismConfig giterminismConfig
}

func (r *FileReader) SetGiterminismConfig(giterminismConfig giterminismConfig) {
	r.giterminismConfig = giterminismConfig
}

func NewFileReader(sharedOptions sharedOptions) FileReader {
	return FileReader{sharedOptions: sharedOptions}
}

type giterminismConfig interface {
	IsUncommittedConfigAccepted() bool
	IsUncommittedConfigTemplateFileAccepted(relPath string) (bool, error)
	IsUncommittedConfigGoTemplateRenderingFileAccepted(relPath string) (bool, error)
	IsUncommittedDockerfileAccepted(relPath string) (bool, error)
	IsUncommittedDockerignoreAccepted(relPath string) (bool, error)
	IsUncommittedHelmFileAccepted(relPath string) (bool, error)
}

type sharedOptions interface {
	ProjectDir() string
	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	LooseGiterminism() bool
}
