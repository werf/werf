package global_warnings

import (
	"context"

	"github.com/werf/logboek"
)

var (
	GlobalWarningLines []string
)

func PrintGlobalWarnings(ctx context.Context) {
	for _, line := range GlobalWarningLines {
		printGlobalWarningLn(ctx, line)
	}
}

func GlobalWarningLn(ctx context.Context, line string) {
	GlobalWarningLines = append(GlobalWarningLines, line)
	printGlobalWarningLn(ctx, line)
}

func printGlobalWarningLn(ctx context.Context, line string) {
	logboek.Context(ctx).Error().LogF("WARNING: %s\n", line)
}
