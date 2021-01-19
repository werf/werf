package inspector

import (
	"fmt"

	"github.com/werf/werf/pkg/giterminism_manager/errors"
)

type ExternalDependencyFoundError error

func NewExternalDependencyFoundError(msg string) ExternalDependencyFoundError {
	return ExternalDependencyFoundError(errors.NewError(fmt.Sprintf("the configuration with external dependency found in the werf config: %s", msg)))
}
