package common

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/util/option"
)

const DefaultBuildParallelTasksLimit = 5

const DefaultCleanupParallelTasksLimit = 10

func SetupParallelOptions(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	SetupParallel(cmdData, cmd)
	SetupParallelTasksLimit(cmdData, cmd, defaultValue)
}

func SetupParallel(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Parallel = new(bool)
	cmd.Flags().BoolVarP(cmdData.Parallel, "parallel", "p", util.GetBoolEnvironmentDefaultTrue("WERF_PARALLEL"), "Run in parallel (default $WERF_PARALLEL or true)")
}

func GetParallel(cmdData *CmdData) bool {
	return *cmdData.Parallel // delegate non-nil value to setup function
}

func GetParallelTasksLimit(cmdData *CmdData) int64 {
	v := *cmdData.ParallelTasksLimit // delegate non-nil value to setup function
	return lo.Ternary(v > 0, v, -1)
}

func SetupParallelTasksLimit(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	parsedValue, err := util.GetInt64EnvVar("WERF_PARALLEL_TASKS_LIMIT")
	if err != nil {
		panic(fmt.Sprintf("unexpected WERF_PARALLEL_TASKS_LIMIT value: %v", err))
	}
	finalDefaultValue := option.PtrValueOrDefault(parsedValue, defaultValue)

	cmdData.ParallelTasksLimit = new(int64)
	cmd.Flags().Int64VarP(cmdData.ParallelTasksLimit, "parallel-tasks-limit", "", finalDefaultValue, "Parallel tasks limit, set -1 to remove the limitation (default $WERF_PARALLEL_TASKS_LIMIT or 5)")
}
