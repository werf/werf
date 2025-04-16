package exec

import (
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/logging"
	exec2 "github.com/werf/werf/v2/pkg/util/exec"
	"github.com/werf/werf/v2/pkg/util/option"
	"github.com/werf/werf/v2/pkg/werf"
)

// Detach executes werf binary in new detached process.
// The detached process will continue to work after termination of parent process.
func Detach(ctx context.Context, args, envs []string) error {
	name := option.ValueOrDefault(os.Getenv("WERF_ORIGINAL_EXECUTABLE"), os.Args[0])

	env := slices.Concat(envs, os.Environ(), []string{"_WERF_BACKGROUND_MODE_ENABLED=1"})
	env = lo.Uniq(env)

	env = lo.Filter(env, func(item string, _ int) bool {
		return !strings.HasPrefix(item, "WERF_ENABLE_PROCESS_EXTERMINATOR=")
	})

	outStream, errStream, err := logging.BackgroundStreams(werf.GetServiceDir())
	if err != nil {
		return err
	}

	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = outStream
	cmd.Stderr = errStream

	cmd = exec2.MakeDetachable(cmd)

	if err := cmd.Start(); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to start background process: %s\n", err)
		return nil
	}

	if err := cmd.Process.Release(); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to detach background process: %s\n", err)
	}

	return nil
}
