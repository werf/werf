package inspector

import (
	"fmt"

	"github.com/werf/werf/v3/pkg/giterminism_manager/errors"
)

func NewExternalDependencyFoundError(msg string) error {
	return errors.NewError(fmt.Sprintf("the configuration with potential external dependency found in the werf config: %s", msg))
}
