package git_repo

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/werf/werf/pkg/determinism_inspector"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/werf/werf/pkg/util"
)

func newHash(s string) (plumbing.Hash, error) {
	var h plumbing.Hash

	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}

	copy(h[:], b)
	return h, nil
}

func ReadGitRepoFileAndCompareWithProjectFile(ctx context.Context, localGitRepo *Local, commit, projectDir, relPath string) ([]byte, error) {
	fileRepoPath := filepath.ToSlash(relPath)

	repoData, err := localGitRepo.ReadFile(ctx, commit, fileRepoPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q from the local git repo commit %s: %s", fileRepoPath, commit, err)
	}

	isDataIdentical, err := compareGitRepoFileWithProjectFile(repoData, projectDir, relPath)
	if err != nil {
		return nil, fmt.Errorf("error comparing repo file %q with the local project file: %s", fileRepoPath, err)
	}

	if !isDataIdentical {
		if err := determinism_inspector.ReportUncommittedFile(ctx, relPath); err != nil {
			return nil, err
		}
	}

	return repoData, err
}

func compareGitRepoFileWithProjectFile(repoFileData []byte, projectDir, relPath string) (bool, error) {
	var localData []byte
	absPath := filepath.Join(projectDir, relPath)
	exist, err := util.FileExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check file existence: %s", err)
	} else if exist {
		localData, err = ioutil.ReadFile(absPath)
		if err != nil {
			return false, fmt.Errorf("unable to read file: %s", err)
		}
	}

	isDataIdentical := bytes.Equal(repoFileData, localData)
	localDataWithForcedUnixLineBreak := bytes.ReplaceAll(localData, []byte("\r\n"), []byte("\n"))
	if !isDataIdentical {
		isDataIdentical = bytes.Equal(repoFileData, localDataWithForcedUnixLineBreak)
	}

	return isDataIdentical, nil
}
