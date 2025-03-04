package exec

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	exec2 "github.com/werf/werf/v2/pkg/util/exec"
	"github.com/werf/werf/v2/pkg/util/option"
)

// Detach executes werf binary in new detached process.
// The detached process will continue to work after termination of parent process.
func Detach(ctx context.Context, args []string) error {
	name := option.ValueOrDefault(os.Getenv("WERF_ORIGINAL_EXECUTABLE"), os.Args[0])

	env := append(os.Environ(), "_WERF_BACKGROUND_MODE_ENABLED=1")

	env = lo.Filter(env, func(item string, _ int) bool {
		return !strings.HasPrefix(item, "WERF_ENABLE_PROCESS_EXTERMINATOR=")
	})

	cmd := exec.Command(name, args...)
	cmd.Env = env

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
