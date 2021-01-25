package inspector

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/giterminism_manager/errors"
)

func NewExternalDependencyFoundError(msg string) error {
	return errors.NewError(fmt.Sprintf("the configuration with external dependency found in the werf config: %s", msg))
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

	return errors.NewError(errorMsg)
}

func prepareListOfFilesString(paths []string) string {
	var result string
	for _, path := range paths {
		result += " - " + filepath.ToSlash(path) + "\n"
	}

	return strings.TrimSuffix(result, "\n")
}
