package exec

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/logging"
	utilExec "github.com/werf/werf/v2/pkg/util/exec"
	"github.com/werf/werf/v2/pkg/util/option"
	"github.com/werf/werf/v2/pkg/werf"
)

// Detach executes werf binary in new detached process.
// The detached process will continue to work after termination of parent process.
func Detach(ctx context.Context, args, envs []string) error {
	env := slices.Concat(envs, os.Environ(), []string{"_WERF_BACKGROUND_MODE_ENABLED=1"})
	env = lo.Uniq(env)

	env = lo.Filter(env, func(item string, _ int) bool {
		return !strings.HasPrefix(item, "WERF_ENABLE_PROCESS_EXTERMINATOR=")
	})

	outStream, errStream, err := logging.BackgroundStreams(werf.GetServiceDir())
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	name := option.ValueOrDefault(os.Getenv("WERF_ORIGINAL_EXECUTABLE"), execPath)

	cmd := CommandContextCancellation(ctx, name, args...)
	cmd.Env = env
	cmd.Stdout = outStream
	cmd.Stderr = errStream

	cmd = utilExec.MakeDetachable(cmd)

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Process.Release(); err != nil {
		return err
	}

	return nil
}
