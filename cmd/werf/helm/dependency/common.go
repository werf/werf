package dependency

import (
	"fmt"
	"strings"
)

func isNoRepositoryDefinitionError(err error) bool {
	return strings.HasPrefix(err.Error(), "no repository definition for")
}

func processNoRepositoryDefinitionError(err error) error {
	return fmt.Errorf(strings.Replace(err.Error(), "helm repo add", "werf helm repo add", -1))
}
