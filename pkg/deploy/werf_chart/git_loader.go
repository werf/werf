package werf_chart

import (
	"context"
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/git_repo"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func LoadFilesFromGit(ctx context.Context, localGitRepo *git_repo.Local, projectDir, loadDir string) ([]*loader.BufferedFile, error) {
	commit, err := localGitRepo.HeadCommit(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit projectDir=%s loadDir=%s commit=%s\n", loadDir, projectDir, commit)

	repoPaths, err := localGitRepo.GetFilePathList(commit)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo paths for commit %s: %s", commit, err)
	}

	relativeLoadDir := util.GetRelativeToBaseFilepath(projectDir, loadDir)

	var res []*loader.BufferedFile

	for _, repoPath := range repoPaths {
		if util.IsSubpathOfBasePath(relativeLoadDir, repoPath) {
			if data, err := localGitRepo.ReadFile(commit, repoPath); err != nil {
				return nil, fmt.Errorf("unable to read file %s for commit %s from local git repo: %s", repoPath, commit, err)
			} else {
				logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit commit=%s loaded file %s:\n%s\n", commit, repoPath, data)

				res = append(res, &loader.BufferedFile{Name: util.GetRelativeToBaseFilepath(relativeLoadDir, repoPath), Data: data})
			}
		}
	}

	return res, nil
}
