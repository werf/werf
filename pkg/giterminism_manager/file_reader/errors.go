package file_reader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/giterminism_manager/errors"
)

type FilesNotFoundInProjectDirectoryError struct {
	error
}
type FilesNotFoundInTheProjectGitRepositoryError struct {
	error
}
type UncommittedFilesError struct {
	error
}
type UncommittedFilesChangesError struct {
	error
}

func isUncommittedFilesChangesError(err error) bool {
	switch err.(type) {
	case UncommittedFilesChangesError:
		return true
	default:
		return false
	}
}

func (r FileReader) NewFilesNotFoundInProjectDirectoryError(relPaths ...string) error {
	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q not found in the project directory", filepath.ToSlash(relPaths[0]))
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files not found in the project directory:\n\n%s", prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	return FilesNotFoundInProjectDirectoryError{errors.NewError(errorMsg)}
}

func (r FileReader) NewFilesNotFoundInProjectGitRepositoryError(relPaths ...string) error {
	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q not found in the project git repository", filepath.ToSlash(relPaths[0]))
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files not found in the project git repository:\n\n%s", prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	return FilesNotFoundInTheProjectGitRepositoryError{errors.NewError(errorMsg)}
}

func (r FileReader) NewUncommittedFilesError(relPaths ...string) error {
	return UncommittedFilesError{r.newUncommittedFilesErrorBase("", relPaths...)}
}

func (r FileReader) NewUncommittedFilesChangesError(relPaths ...string) error {
	return UncommittedFilesChangesError{UncommittedFilesError{r.newUncommittedFilesErrorBase("changes", relPaths...)}}
}

func (r FileReader) newUncommittedFilesErrorBase(specificFileProperty string, relPaths ...string) error {
	if specificFileProperty != "" {
		specificFileProperty = " " + specificFileProperty
	}

	expectedAction := "must be committed"
	if r.sharedOptions.Dev() {
		expectedAction = "must be staged"
	}

	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q%s %s", filepath.ToSlash(relPaths[0]), specificFileProperty, expectedAction)
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files%s %s:\n\n%s", specificFileProperty, expectedAction, prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	if r.sharedOptions.Dev() {
		errorMsg = fmt.Sprintf(`%s

To stage the changes use the following command: "git add %s".`, errorMsg, strings.Join(formatFilePathList(relPaths), " "))
	} else {
		errorMsg = fmt.Sprintf(`%s

You might also be interested in developer mode (activated with --dev option) that allows you to work with staged changes without doing redundant commits. Just use "git add <file>..." to include the changes that should be used.`, errorMsg)
	}

	return UncommittedFilesChangesError{errors.NewError(errorMsg)}
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
