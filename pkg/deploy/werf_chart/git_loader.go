package werf_chart

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/werf/logboek"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/git_repo"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func GiterministicFilesLoader(ctx context.Context, localGitRepo *git_repo.Local, projectDir, loadDir string) ([]*loader.BufferedFile, error) {
	var res []*loader.BufferedFile
	var lock *chart.Lock

	if gitFiles, err := LoadFilesFromGit(ctx, localGitRepo, projectDir, loadDir); err != nil {
		return nil, err
	} else {
		for _, f := range gitFiles {
			switch {
			case f.Name == "Chart.lock":
				lock = new(chart.Lock)
				if err := yaml.Unmarshal(f.Data, &lock); err != nil {
					return nil, errors.Wrap(err, "cannot load Chart.lock")
				}
				break
			case f.Name == "requirements.lock":
				lock = new(chart.Lock)
				if err := yaml.Unmarshal(f.Data, &lock); err != nil {
					return nil, errors.Wrap(err, "cannot load requirements.lock")
				}
				break
			}
		}

		res = gitFiles
	}

	localFiles, err := loader.GetFilesFromLocalFilesystem(loadDir)
	if err != nil {
		return nil, err
	}

	if lock != nil {
		localSubchartsFiles := make(map[string][]*loader.BufferedFile)

		for _, f := range localFiles {
			switch {
			case strings.HasPrefix(f.Name, "charts/"):
				fname := strings.TrimPrefix(f.Name, "charts/")
				cname := strings.SplitN(fname, "/", 2)[0]
				localSubchartsFiles[cname] = append(localSubchartsFiles[cname], f)
				logboek.Context(ctx).Debug().LogF("-- GiterministicFilesLoader: local subchart %q: found file %q\n", cname, f.Name)
			}
		}

		for _, dep := range lock.Dependencies {
			fullDepName := fmt.Sprintf("%s-%s.tgz", dep.Name, dep.Version)
			if files, hasKey := localSubchartsFiles[fullDepName]; hasKey {
				logboek.Context(ctx).Debug().LogF("-- GiterministicFilesLoader: using subchart %q from the local filesystem\n", fullDepName)
				res = append(res, files...)
			}
		}
	}

	return res, nil
}

func LoadFilesFromGit(ctx context.Context, localGitRepo *git_repo.Local, projectDir, loadDir string) ([]*loader.BufferedFile, error) {
	commit, err := localGitRepo.HeadCommit(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	logboek.Context(ctx).Debug().LogF("-- LoadFilesFromGit projectDir=%s loadDir=%s commit=%s\n", loadDir, projectDir, commit)

	repoPaths, err := localGitRepo.GetCommitFilePathList(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("unable to get local repo paths for commit %s: %s", commit, err)
	}

	// FIXME: .helmignore

	relativeLoadDir := util.GetRelativeToBaseFilepath(projectDir, loadDir)

	if isSymlink, linkDest, err := localGitRepo.CheckAndReadCommitSymlink(ctx, relativeLoadDir, commit); err != nil {
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
				res = append(res, &loader.BufferedFile{Name: filepath.ToSlash(util.GetRelativeToBaseFilepath(relativeLoadDir, repoPath)), Data: d})
			}
		}
	}

	return res, nil
}
