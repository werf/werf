package cleanup

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/build"
)

type CommonProjectOptions struct {
	ProjectName   string
	CommonOptions CommonOptions
}

func projectCleanup(options CommonProjectOptions) error {
	filterSet := projectFilterSet(options)
	filterSet.Add("dangling", "true")
	if err := werfImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	if err := werfContainersFlushByFilterSet(projectFilterSet(options), options.CommonOptions); err != nil {
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
	filterSet.Add("label", werfLabel(options))
	return filterSet
}

func werfLabel(options CommonProjectOptions) string {
	return fmt.Sprintf("werf=%s", options.ProjectName)
}

func stageCacheReference(options CommonProjectOptions) string {
	return fmt.Sprintf(build.LocalDimgstageImageNameFormat, options.ProjectName)
}
