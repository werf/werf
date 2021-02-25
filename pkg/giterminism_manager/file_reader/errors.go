package file_reader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/giterminism_manager/errors"
)

type UncommittedFilesError struct {
	error
}
type FileNotAcceptedError struct {
	error
}
type FileNotFoundInProjectDirectoryError struct {
	error
}
type FileNotFoundInProjectRepositoryError struct {
	error
}

func IsFileNotFoundInProjectDirectoryError(err error) bool {
	switch err.(type) {
	case FileNotFoundInProjectDirectoryError:
		return true
	default:
		return false
	}
}

func (r FileReader) NewFileNotFoundInProjectDirectoryError(relPath string) error {
	return FileNotFoundInProjectDirectoryError{errors.NewError(fmt.Sprintf("the file %q not found in the project directory", filepath.ToSlash(relPath)))}
}

func (r FileReader) NewFileNotFoundInProjectRepositoryError(relPath string) error {
	return FileNotFoundInProjectRepositoryError{errors.NewError(fmt.Sprintf("the file %q not found in the project git repository", filepath.ToSlash(relPath)))}
}

func (r FileReader) NewSubmoduleAddedAndNotCommittedError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the added submodule %q must be committed", filepath.ToSlash(submodulePath))
	return r.newUncommittedFilesErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewSubmoduleDeletedError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the deleted submodule %q must be committed", filepath.ToSlash(submodulePath))
	return r.newUncommittedFilesErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewSubmoduleHasUntrackedChangesError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the submodule %q has untracked changes that must be discarded or committed (do not forget to push new changes to the submodule remote)", filepath.ToSlash(submodulePath))
	return UncommittedFilesError{fmt.Errorf("%s", errorMsg)}
}

func (r FileReader) NewSubmoduleHasUncommittedChangesError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the submodule %q has uncommitted changes that must be discarded or committed (do not forget to push new changes to the submodule remote)", filepath.ToSlash(submodulePath))
	return UncommittedFilesError{fmt.Errorf("%s", errorMsg)}
}

func (r FileReader) NewSubmoduleCommitChangedError(submodulePath string) error {
	expectedAction := "must be committed"
	if r.sharedOptions.Dev() {
		expectedAction = "must be staged"
	}
	errorMsg := fmt.Sprintf("the submodule %q is not clean and %s. Do not forget to push the current commit to the submodule remote if this commit exists only locally", filepath.ToSlash(submodulePath), expectedAction)

	return r.newUncommittedFilesErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewUncommittedFilesError(relPaths ...string) error {
	expectedAction := "must be committed"
	if r.sharedOptions.Dev() {
		expectedAction = "must be staged"
	}

	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q %s", filepath.ToSlash(relPaths[0]), expectedAction)
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files %s:\n\n%s", expectedAction, prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	return UncommittedFilesError{r.newUncommittedFilesErrorBase(errorMsg, strings.Join(formatFilePathList(relPaths), " "))}
}

func (r FileReader) newUncommittedFilesErrorBase(errorMsg string, gitAddArg string) error {
	if r.sharedOptions.Dev() {
		errorMsg = fmt.Sprintf(`%s

To stage the changes use the following command: "git add %s".`, errorMsg, gitAddArg)
	} else {
		errorMsg = fmt.Sprintf(`%s

You might also be interested in developer mode (activated with --dev option) that allows you to work with staged changes without doing redundant commits. Just use "git add <file>..." to include the changes that should be used.`, errorMsg)
	}

	return UncommittedFilesError{errors.NewError(errorMsg)}
}

func prepareListOfFilesString(paths []string) string {
	return " - " + strings.Join(formatFilePathList(paths), "\n - ")
}

func formatFilePathList(paths []string) []string {
	var result []string
	for _, path := range paths {
		result = append(result, filepath.ToSlash(path))
	}

	return result
}
