package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/logboek"

	"github.com/bmatcuk/doublestar"
)

func (r FileReader) relativeToGitWorkingDir() string {
	workingDir := r.sharedOptions.ProjectDir() // FIXME: rename project-dir to working dir and work-tree-dir to project-git-work-tree or project-git or project-work-tree or ...
	return util.GetRelativeToBaseFilepath(r.sharedOptions.LocalGitRepo().WorkTreeDir, workingDir)
}

func (r FileReader) relativeToGitPath(relPath string) string {
	return filepath.Join(r.relativeToGitWorkingDir(), relPath)
}

func (r FileReader) isCommitDirectoryExist(ctx context.Context, relPath string) (bool, error) {
	exist, err := r.sharedOptions.LocalGitRepo().IsCommitDirectoryExists(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
	if err != nil {
		err := fmt.Errorf(
			"unable to check existence of directory %s in the project git repo commit %s: %s",
			relPath, r.sharedOptions.HeadCommit(), err,
		)
		return false, err
	}

	return exist, nil
}

func (r FileReader) isCommitFileExist(ctx context.Context, relPath string) (bool, error) {
	logboek.Context(ctx).Debug().LogF("-- giterminism_manager.FileReader.isCommitFileExist relPath=%q\n", relPath)

	exist, err := r.sharedOptions.LocalGitRepo().IsCommitFileExists(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
	if err != nil {
		err := fmt.Errorf(
			"unable to check existence of file %s in the project git repo commit %s: %s",
			relPath, r.sharedOptions.HeadCommit(), err,
		)
		return false, err
	}

	return exist, nil
}

func (r FileReader) commitFilesGlob(ctx context.Context, pattern string) ([]string, error) {
	var result []string

	commitPathList, err := r.sharedOptions.LocalGitRepo().GetCommitFilePathList(ctx, r.sharedOptions.HeadCommit())
	if err != nil {
		return nil, fmt.Errorf("unable to get files list from local git repo: %s", err)
	}

	pattern = filepath.ToSlash(pattern)
	for _, relToGitFilepath := range commitPathList {
		relToGitPath := filepath.ToSlash(relToGitFilepath)
		relPath := util.GetRelativeToBaseFilepath(r.relativeToGitWorkingDir(), relToGitPath)

		if matched, err := doublestar.Match(pattern, relPath); err != nil {
			return nil, err
		} else if matched {
			result = append(result, relPath)
		}
	}

	return result, nil
}

func (r FileReader) readCommitFile(ctx context.Context, relPath string, handleUncommittedChangesFunc func(context.Context, string) error) ([]byte, error) {
	logboek.Context(ctx).Debug().LogF("-- giterminism_manager.FileReader.readCommitFile relPath=%q\n", relPath)

	fileRepoPath := r.relativeToGitPath(filepath.ToSlash(relPath))

	repoData, err := r.sharedOptions.LocalGitRepo().ReadCommitFile(ctx, r.sharedOptions.HeadCommit(), fileRepoPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q from the local git repo commit %s: %s", relPath, r.sharedOptions.HeadCommit(), err)
	}

	if handleUncommittedChangesFunc == nil {
		return repoData, nil
	}

	isDataIdentical, err := r.compareFileData(relPath, repoData)
	if err != nil {
		return nil, fmt.Errorf("unable to compare commit file %q with the local project file: %s", relPath, err)
	}

	if !isDataIdentical {
		if err := handleUncommittedChangesFunc(ctx, filepath.FromSlash(relPath)); err != nil {
			return nil, err
		}
	}

	return repoData, err
}

func (r FileReader) compareFileData(relPath string, data []byte) (bool, error) {
	var fileData []byte
	if exist, err := r.isFileExist(relPath); err != nil {
		return false, err
	} else if exist {
		fileData, err = r.readFile(relPath)
		if err != nil {
			return false, err
		}
	}

	isDataIdentical := bytes.Equal(fileData, data)
	fileDataWithForcedUnixLineBreak := bytes.ReplaceAll(fileData, []byte("\r\n"), []byte("\n"))
	if !isDataIdentical {
		isDataIdentical = bytes.Equal(fileDataWithForcedUnixLineBreak, data)
	}

	return isDataIdentical, nil
}
