package inspector

import (
	"context"

	"github.com/werf/werf/pkg/git_repo"
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
	IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error)
	IsConfigStapelFromLatestAccepted() bool
	IsConfigStapelGitBranchAccepted() bool
	IsConfigStapelMountBuildDirAccepted() bool
	IsConfigStapelMountFromPathAccepted(fromPath string) (bool, error)
	IsConfigDockerfileContextAddFileAccepted(relPath string) (bool, error)
}

type fileReader interface {
	HandleValidateSubmodulesErr(err error) error
	ExtraCheckFilesModifiedLocally(ctx context.Context, relPath ...string) error
}

type sharedOptions interface {
	RelativeToGitProjectDir() string
	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	LooseGiterminism() bool
	Dev() bool
}
