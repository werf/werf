package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/util"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
)

func (r FileReader) LocateChart(ctx context.Context, name string, settings *cli.EnvSettings) (string, error) {
	commit, err := r.manager.LocalGitRepo().HeadCommit(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	if exists, err := r.manager.LocalGitRepo().IsCommitDirectoryExists(ctx, name, commit); err != nil {
		return "", fmt.Errorf("error checking existence of %q in the local git repo commit %s: %s", name, commit, err)
	} else if exists {
		return name, nil
	}
	return "", fmt.Errorf("chart path %q not found in the local git repo commit %s", name, commit)
}

func (r FileReader) ReadChartFile(ctx context.Context, filePath string) ([]byte, error) {
	commit, err := r.manager.LocalGitRepo().HeadCommit(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	relativeFilePath := util.GetRelativeToBaseFilepath(r.manager.ProjectDir(), filePath)

	return git_repo.ReadCommitFileAndCompareWithProjectFile(ctx, *r.manager.LocalGitRepo(), commit, r.manager.ProjectDir(), relativeFilePath)
}

func (r FileReader) LoadChartDir(ctx context.Context, dir string) ([]*chart.ChartExtenderBufferedFile, error) {
	commit, err := r.manager.LocalGitRepo().HeadCommit(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit dir=%s projectDir=%s commit=%s\n", dir, r.manager.ProjectDir(), commit)

	repoPaths, err := r.manager.LocalGitRepo().GetCommitFilePathList(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo paths for commit %s: %s", commit, err)
	}

	// TODO: Add .helmignore

	relativeLoadDir := util.GetRelativeToBaseFilepath(r.manager.ProjectDir(), dir)

	if isSymlink, linkDest, err := r.manager.LocalGitRepo().CheckAndReadCommitSymlink(ctx, relativeLoadDir, commit); err != nil {
		return nil, fmt.Errorf("error checking %s is symlink in the local git repo commit %s: %s", relativeLoadDir, commit, err)
	} else if isSymlink {
		logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit: load dir %q is symlink to %q\n", relativeLoadDir, linkDest)
		relativeLoadDir = string(linkDest)
	}

	var res []*chart.ChartExtenderBufferedFile

	for _, repoPath := range repoPaths {
		if util.IsSubpathOfBasePath(relativeLoadDir, repoPath) {
			if d, err := git_repo.ReadCommitFileAndCompareWithProjectFile(ctx, *r.manager.LocalGitRepo(), commit, r.manager.ProjectDir(), repoPath); err != nil {
				return nil, err
			} else {
				logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit commit=%s loaded file %s:\n%s\n", commit, repoPath, d)
				res = append(res, &chart.ChartExtenderBufferedFile{Name: filepath.ToSlash(util.GetRelativeToBaseFilepath(relativeLoadDir, repoPath)), Data: d})
			}
		}
	}

	return res, nil
}
