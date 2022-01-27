package file_reader

import (
	"os"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/path_matcher"
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
	UncommittedConfigTemplateFilePathMatcher() path_matcher.PathMatcher
	UncommittedConfigGoTemplateRenderingFilePathMatcher() path_matcher.PathMatcher
	IsUncommittedDockerfileAccepted(relPath string) bool
	IsUncommittedDockerignoreAccepted(relPath string) bool
	UncommittedHelmFilePathMatcher() path_matcher.PathMatcher
}

type sharedOptions interface {
	ProjectDir() string
	RelativeToGitProjectDir() string
	LocalGitRepo() git_repo.GitRepo
	HeadCommit() string
	LooseGiterminism() bool
	Dev() bool
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_GITERMINISM_MANAGER") == "1"
}
