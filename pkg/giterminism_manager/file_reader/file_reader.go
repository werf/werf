package file_reader

import (
	"context"
	"os"
	"path/filepath"

	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type FileReader struct {
	sharedOptions     sharedOptions
	giterminismConfig giterminismConfig
	fileSystem        fileSystem
}

func (r *FileReader) SetGiterminismConfig(giterminismConfig giterminismConfig) {
	r.giterminismConfig = giterminismConfig
}

func (r *FileReader) SetFileSystemLayer(fileSystemLayer fileSystem) {
	r.fileSystem = fileSystemLayer
}

func NewFileReader(sharedOptions sharedOptions) FileReader {
	return FileReader{
		sharedOptions: sharedOptions,
		fileSystem:    newFileSystemOperator(),
	}
}

//go:generate mockgen -source file_reader.go -package file_reader_test -destination file_reader_deps_mock_test.go

//go:generate mockgen -package file_reader_test -destination file_reader_deps_fileinfo_mock_test.go os FileInfo

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
	HeadCommit(ctx context.Context) string
	LooseGiterminism() bool
	Dev() bool
}

type fileSystem interface {
	ReadFile(filename string) ([]byte, error)
	Readlink(name string) (string, error)
	IsNotExist(err error) bool
	Walk(root string, fn filepath.WalkFunc) error
	Lstat(name string) (os.FileInfo, error)
	FileExists(p string) (bool, error)
	DirExists(path string) (bool, error)
	RegularFileExists(path string) (bool, error)
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_GITERMINISM_MANAGER") == "1"
}

func applyDebugToLogboek(options types.LogBlockOptionsInterface) {
	if !debug() {
		options.Mute()
	}
}
