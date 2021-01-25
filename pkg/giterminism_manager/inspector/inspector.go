package inspector

import "github.com/werf/werf/pkg/git_repo"

type Inspector struct {
	giterminismConfig giterminismConfig
	sharedOptions     sharedOptions
}

func NewInspector(giterminismConfig giterminismConfig, sharedOptions sharedOptions) Inspector {
	return Inspector{giterminismConfig: giterminismConfig, sharedOptions: sharedOptions}
}

type giterminismConfig interface {
	IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error)
	IsConfigStapelFromLatestAccepted() bool
	IsConfigStapelGitBranchAccepted() bool
	IsConfigStapelMountBuildDirAccepted() bool
	IsConfigStapelMountFromPathAccepted(fromPath string) (bool, error)
	IsConfigDockerfileContextAddFileAccepted(relPath string) (bool, error)
}

type sharedOptions interface {
	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	LooseGiterminism() bool
	Dev() bool
}
