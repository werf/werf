package cleanup

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"
)

type CommonProjectOptions struct {
	ProjectName   string        `json:"project_name"`
	CommonOptions CommonOptions `json:"common_options"`
}

func dappProjectCleanup(options CommonProjectOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", dappLabel(options))
	filterSet.Add("dangling", "true")
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	filterSet = filters.NewArgs()
	filterSet.Add("label", dappLabel(options))
	if err := dappContainersFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func dappLabel(options CommonProjectOptions) string {
	return fmt.Sprintf("dapp=%s", options.ProjectName)
}

func stageCacheReference(options CommonProjectOptions) string {
	return fmt.Sprintf("dimgstage-%s", options.ProjectName)
}
