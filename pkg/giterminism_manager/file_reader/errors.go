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

func NewFilesNotFoundInProjectDirectoryError(relPaths ...string) error {
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

func NewFilesNotFoundInProjectGitRepositoryError(relPaths ...string) error {
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

func NewUncommittedFilesError(relPaths ...string) error {
	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q must be committed", filepath.ToSlash(relPaths[0]))
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files must be committed:\n\n%s", prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	return UncommittedFilesError{errors.NewError(errorMsg)}
}

func NewUncommittedFilesChangesError(relPaths ...string) error {
	var errorMsg string
	if len(relPaths) == 1 {
		errorMsg = fmt.Sprintf("the file %q changes must be committed", filepath.ToSlash(relPaths[0]))
	} else if len(relPaths) > 1 {
		errorMsg = fmt.Sprintf("the following files changes must be committed:\n\n%s", prepareListOfFilesString(relPaths))
	} else {
		panic("unexpected condition")
	}

	return UncommittedFilesChangesError{errors.NewError(errorMsg)}
}

func prepareListOfFilesString(paths []string) string {
	var result string
	for _, path := range paths {
		result += " - " + filepath.ToSlash(path) + "\n"
	}

	return strings.TrimSuffix(result, "\n")
}
