package true_git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/werf/werf/pkg/util"
)

func UpwardLookupAndVerifyWorkTree(lookupPath string) (bool, string, error) {
	lookupPath = util.GetAbsoluteFilepath(lookupPath)

	for {
		dotGitPath := filepath.Join(lookupPath, git.GitDirName)

		if _, err := os.Stat(dotGitPath); os.IsNotExist(err) {
			if lookupPath != filepath.Dir(lookupPath) {
				lookupPath = filepath.Dir(lookupPath)
				continue
			}
		} else if err != nil {
			return false, "", fmt.Errorf("error accessing %q: %s", dotGitPath, err)
		} else if isValid, err := IsValidGitDir(dotGitPath); err != nil {
			return false, "", err
		} else if isValid {
			return true, lookupPath, nil
		}

		break
	}

	return false, "", nil
}

func IsValidWorkTree(workTree string) (bool, error) {
	return IsValidGitDir(filepath.Join(workTree, git.GitDirName))
}

func IsValidGitDir(gitDir string) (bool, error) {
	gitArgs := []string{"--git-dir", gitDir, "rev-parse"}

	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.HasPrefix(string(output), "fatal: not a git repository: ") {
			return false, nil
		}
		return false, fmt.Errorf("%v failed: %s:\n%s", strings.Join(append([]string{"git"}, gitArgs...), " "), err, output)
	}

	return true, nil
}
