package common

import (
	"fmt"

	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
)

func GetConveyorOptions(commonCmdData *CmdData) build.ConveyorOptions {
	return build.ConveyorOptions{
		LocalGitRepoVirtualMergeOptions: stage.VirtualMergeOptions{
			VirtualMerge:           *commonCmdData.VirtualMerge,
			VirtualMergeFromCommit: *commonCmdData.VirtualMergeFromCommit,
			VirtualMergeIntoCommit: *commonCmdData.VirtualMergeIntoCommit,
		},
	}
}

func GetConveyorOptionsWithParallel(commonCmdData *CmdData, buildStagesOptions build.BuildOptions) (build.ConveyorOptions, error) {
	conveyorOptions := GetConveyorOptions(commonCmdData)
	conveyorOptions.Parallel = !(buildStagesOptions.ImageBuildOptions.IntrospectAfterError || buildStagesOptions.ImageBuildOptions.IntrospectBeforeError || len(buildStagesOptions.Targets) != 0) && *commonCmdData.Parallel

	parallelTasksLimit, err := GetParallelTasksLimit(commonCmdData)
	if err != nil {
		return conveyorOptions, fmt.Errorf("getting parallel tasks limit failed: %s", err)
	}

	conveyorOptions.ParallelTasksLimit = parallelTasksLimit

	return conveyorOptions, nil
}

func GetBuildOptions(commonCmdData *CmdData, werfConfig *config.WerfConfig) (buildOptions build.BuildOptions, err error) {
	introspectOptions, err := GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return buildOptions, err
	}

	reportFormat, err := GetReportFormat(commonCmdData)
	if err != nil {
		return buildOptions, err
	}

	buildOptions = build.BuildOptions{
		ImageBuildOptions: container_runtime.BuildOptions{
			IntrospectAfterError:  *commonCmdData.IntrospectAfterError,
			IntrospectBeforeError: *commonCmdData.IntrospectBeforeError,
		},
		IntrospectOptions: introspectOptions,
		ReportPath:        *commonCmdData.ReportPath,
		ReportFormat:      reportFormat,
	}

	return buildOptions, nil
}
