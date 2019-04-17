package cleaning

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/build"
)

const (
	localStagesStorage string = ":local"
)

type CommonProjectOptions struct {
	ProjectName   string
	CommonOptions CommonOptions
}

func projectImageStageFilterSet(options CommonProjectOptions) filters.Args {
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
	return fmt.Sprintf(build.LocalImageStageImageNameFormat, options.ProjectName)
}
