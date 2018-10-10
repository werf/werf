package cleanup

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"
)

type CommonProjectOptions struct {
	ProjectName   string        `json:"project_name"`
	CommonOptions CommonOptions `json:"common_options"`
}

func projectCleanup(options CommonProjectOptions) error {
	filterSet := projectDimgstageFilterSet(options)
	filterSet.Add("dangling", "true")
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	if err := dappContainersFlushByFilterSet(projectFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func projectDimgstageFilterSet(options CommonProjectOptions) filters.Args {
	filterSet := projectFilterSet(options)
	filterSet.Add("reference", stageCacheReference(options))
	return filterSet
}

func projectFilterSet(options CommonProjectOptions) filters.Args {
	filterSet := filters.NewArgs()
	filterSet.Add("label", dappLabel(options))
	return filterSet
}

func dappLabel(options CommonProjectOptions) string {
	return fmt.Sprintf("dapp=%s", options.ProjectName)
}

func stageCacheReference(options CommonProjectOptions) string {
	return fmt.Sprintf("dimgstage-%s", options.ProjectName)
}
