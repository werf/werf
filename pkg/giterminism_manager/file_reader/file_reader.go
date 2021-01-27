package file_reader

import "github.com/werf/werf/pkg/git_repo"

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
	RelativeToGitProjectDir() string
	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	LooseGiterminism() bool
}
