package common

import (
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
)

func GetConveyorOptions(commonCmdData *CmdData) build.ConveyorOptions {
	var parallel bool
	if commonCmdData.Parallel != nil {
		parallel = *commonCmdData.Parallel
	}

	return build.ConveyorOptions{
		LocalGitRepoVirtualMergeOptions: stage.VirtualMergeOptions{
			VirtualMerge:           *commonCmdData.VirtualMerge,
			VirtualMergeFromCommit: *commonCmdData.VirtualMergeFromCommit,
			VirtualMergeIntoCommit: *commonCmdData.VirtualMergeIntoCommit,
		},
		GitUnshallow:         *commonCmdData.GitUnshallow,
		AllowGitShallowClone: *commonCmdData.AllowGitShallowClone,
		Parallel:             parallel,
	}
}
