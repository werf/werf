package filemanager

import (
	"context"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type fakeSharedOptions struct {
	projectDir string
	// localGitRepo stays nil in loose mode, where nothing reaches for git.
	localGitRepo        git_repo.GitRepo
	enforcedGiterminism bool
}

func (o fakeSharedOptions) ProjectDir() string                  { return o.projectDir }
func (o fakeSharedOptions) RelativeToGitProjectDir() string     { return "" }
func (o fakeSharedOptions) LocalGitRepo() git_repo.GitRepo      { return o.localGitRepo }
func (o fakeSharedOptions) HeadCommit(_ context.Context) string { return "head commit" }
func (o fakeSharedOptions) LooseGiterminism() bool              { return !o.enforcedGiterminism }
func (o fakeSharedOptions) Dev() bool                           { return false }

type fakeGiterminismConfig struct {
	// uncommittedHelmFilesRejected makes helm files come from the commit instead of the worktree,
	// which is what enforced giterminism does for a chart without an uncommitted-files exception.
	uncommittedHelmFilesRejected bool
}

func (c fakeGiterminismConfig) IsUncommittedConfigAccepted() bool { return true }

func (c fakeGiterminismConfig) UncommittedConfigTemplateFilePathMatcher() path_matcher.PathMatcher {
	return path_matcher.NewTruePathMatcher()
}

func (c fakeGiterminismConfig) UncommittedConfigGoTemplateRenderingFilePathMatcher() path_matcher.PathMatcher {
	return path_matcher.NewTruePathMatcher()
}

func (c fakeGiterminismConfig) IsUncommittedDockerfileAccepted(_ string) bool   { return true }
func (c fakeGiterminismConfig) IsUncommittedDockerignoreAccepted(_ string) bool { return true }

func (c fakeGiterminismConfig) UncommittedHelmFilePathMatcher() path_matcher.PathMatcher {
	if c.uncommittedHelmFilesRejected {
		return path_matcher.NewFalsePathMatcher()
	}
	return path_matcher.NewTruePathMatcher()
}
