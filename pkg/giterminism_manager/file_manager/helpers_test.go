package filemanager

import (
	"context"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type fakeSharedOptions struct {
	projectDir string
}

func (o fakeSharedOptions) ProjectDir() string                  { return o.projectDir }
func (o fakeSharedOptions) RelativeToGitProjectDir() string     { return "" }
func (o fakeSharedOptions) LocalGitRepo() git_repo.GitRepo      { return nil }
func (o fakeSharedOptions) HeadCommit(_ context.Context) string { return "head commit" }
func (o fakeSharedOptions) LooseGiterminism() bool              { return true }
func (o fakeSharedOptions) Dev() bool                           { return false }

type fakeGiterminismConfig struct{}

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
	return path_matcher.NewTruePathMatcher()
}
