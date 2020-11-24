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

	repoPaths, err := localGitRepo.GetFilePathList(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo paths for commit %s: %s", commit, err)
	}

	relativeLoadDir := util.GetRelativeToBaseFilepath(projectDir, loadDir)

	if isSymlink, linkDest, err := localGitRepo.CheckAndReadSymlink(ctx, relativeLoadDir, commit); err != nil {
		return nil, fmt.Errorf("error checking %s is symlink in the local git repo commit %s: %s", relativeLoadDir, commit, err)
	} else if isSymlink {
		logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit: load dir %q is symlink to %q\n", relativeLoadDir, linkDest)
		relativeLoadDir = string(linkDest)
	}

	var res []*loader.BufferedFile

	for _, repoPath := range repoPaths {
		if util.IsSubpathOfBasePath(relativeLoadDir, repoPath) {
			if d, err := git_repo.ReadGitRepoFileAndCompareWithProjectFile(ctx, localGitRepo, commit, projectDir, repoPath); err != nil {
				return nil, err
			} else {
				logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit commit=%s loaded file %s:\n%s\n", commit, repoPath, d)
				res = append(res, &loader.BufferedFile{Name: util.GetRelativeToBaseFilepath(relativeLoadDir, repoPath), Data: d})
			}
		}
	}

	return res, nil
}
