package file_reader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/giterminism_manager/errors"
)

type (
	UntrackedFilesError                  FileReaderError
	UncommittedFilesError                FileReaderError
	FileNotAcceptedError                 FileReaderError
	FileNotFoundInProjectDirectoryError  FileReaderError
	FileNotFoundInProjectRepositoryError FileReaderError
	FileReaderError                      struct {
		error
	}
)

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
	return r.uncommittedErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewSubmoduleDeletedError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the deleted submodule %q must be committed", filepath.ToSlash(submodulePath))
	return r.uncommittedErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewSubmoduleHasUntrackedChangesError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the submodule %q has untracked changes that must be discarded or committed (do not forget to push new changes to the submodule remote)", filepath.ToSlash(submodulePath))
	return UntrackedFilesError{errors.NewError(errorMsg)}
}

func (r FileReader) NewSubmoduleHasUncommittedChangesError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the submodule %q has uncommitted changes that must be discarded or committed (do not forget to push new changes to the submodule remote)", filepath.ToSlash(submodulePath))
	return UncommittedFilesError{fmt.Errorf("%s", errorMsg)}
}

func (r FileReader) NewSubmoduleCommitChangedError(submodulePath string) error {
	errorMsg := fmt.Sprintf("the submodule %q is not clean and %s. Do not forget to push the current commit to the submodule remote if this commit exists only locally", filepath.ToSlash(submodulePath), r.uncommittedUntrackedExpectedAction())
	return r.uncommittedErrorBase(errorMsg, filepath.ToSlash(submodulePath))
}

func (r FileReader) NewUntrackedFilesError(relPaths ...string) error {
	var errorMsg string
	switch {
	case len(relPaths) == 1:
		errorMsg = fmt.Sprintf("the untracked file %q %s", filepath.ToSlash(relPaths[0]), r.uncommittedUntrackedExpectedAction())
	case len(relPaths) > 1:
		errorMsg = fmt.Sprintf("the following untracked files %s:\n\n%s", r.uncommittedUntrackedExpectedAction(), prepareListOfFilesString(relPaths))
	default:
		panic("unexpected condition")
	}

	return UntrackedFilesError{r.uncommittedErrorBase(errorMsg, strings.Join(formatFilePathList(relPaths), " "))}
}

func (r FileReader) NewUncommittedFilesError(relPaths ...string) error {
	var errorMsg string
	switch {
	case len(relPaths) == 1:
		errorMsg = fmt.Sprintf("the file %q %s", filepath.ToSlash(relPaths[0]), r.uncommittedUntrackedExpectedAction())
	case len(relPaths) > 1:
		errorMsg = fmt.Sprintf("the following files %s:\n\n%s", r.uncommittedUntrackedExpectedAction(), prepareListOfFilesString(relPaths))
	default:
		panic("unexpected condition")
	}

	return r.uncommittedErrorBase(errorMsg, strings.Join(formatFilePathList(relPaths), " "))
}

func (r FileReader) uncommittedErrorBase(errorMsg, gitAddArg string) error {
	if r.sharedOptions.Dev() {
		errorMsg = fmt.Sprintf(`%s

To stage the changes use the following command: "git add %s".`, errorMsg, gitAddArg)
	} else {
		errorMsg = fmt.Sprintf(`%s

You may be interested in the development mode (activated by the --dev option), which allows working with project files without doing redundant commits during debugging and development.`, errorMsg)
	}

	return UncommittedFilesError{errors.NewError(errorMsg)}
}

func (r FileReader) uncommittedUntrackedExpectedAction() string {
	if r.sharedOptions.Dev() {
		return "must be staged"
	} else {
		return "must be committed"
	}
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
