package inspector

import (
	"context"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/path_matcher"
)

type Inspector struct {
	giterminismConfig giterminismConfig
	fileReader        fileReader

	sharedOptions sharedOptions
}

func NewInspector(giterminismConfig giterminismConfig, fileReader fileReader, sharedOptions sharedOptions) Inspector {
	return Inspector{giterminismConfig: giterminismConfig, fileReader: fileReader, sharedOptions: sharedOptions}
}

type giterminismConfig interface {
	IsCustomTagsAccepted() bool
	IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error)
	IsConfigStapelFromLatestAccepted() bool
	IsConfigStapelGitBranchAccepted() bool
	IsConfigStapelMountBuildDirAccepted() bool
	IsConfigStapelMountFromPathAccepted(fromPath string) bool
	IsConfigDockerfileContextAddFileAccepted(relPath string) bool
}

type fileReader interface {
	ValidateStatusResult(ctx context.Context, pathMatcher path_matcher.PathMatcher) error
}

type sharedOptions interface {
	RelativeToGitProjectDir() string
	LocalGitRepo() git_repo.GitRepo
	HeadCommit() string
	LooseGiterminism() bool
	Dev() bool
}
