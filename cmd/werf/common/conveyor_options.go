package common

import (
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
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
