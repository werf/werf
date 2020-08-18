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
		GitUnshallow:         *commonCmdData.GitUnshallow,
		AllowGitShallowClone: *commonCmdData.AllowGitShallowClone,
	}
}

func GetConveyorOptionsWithParallel(commonCmdData *CmdData, buildStagesOptions build.BuildStagesOptions) (build.ConveyorOptions, error) {
	conveyorOptions := GetConveyorOptions(commonCmdData)
	conveyorOptions.Parallel = !(buildStagesOptions.ImageBuildOptions.IntrospectAfterError || buildStagesOptions.ImageBuildOptions.IntrospectBeforeError || len(buildStagesOptions.Targets) != 0) && *commonCmdData.Parallel

	parallelTasksLimit, err := GetParallelTasksLimit(commonCmdData)
	if err != nil {
		return conveyorOptions, fmt.Errorf("getting parallel tasks limit failed: %s", err)
	}

	conveyorOptions.ParallelTasksLimit = parallelTasksLimit

	return conveyorOptions, nil
}

func GetBuildStagesOptions(commonCmdData *CmdData, werfConfig *config.WerfConfig) (build.BuildStagesOptions, error) {
	introspectOptions, err := GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return build.BuildStagesOptions{}, err
	}

	options := build.BuildStagesOptions{
		ImageBuildOptions: container_runtime.BuildOptions{
			IntrospectAfterError:  *commonCmdData.IntrospectAfterError,
			IntrospectBeforeError: *commonCmdData.IntrospectBeforeError,
		},
		IntrospectOptions: introspectOptions,
	}

	return options, nil
}
