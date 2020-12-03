package global_warnings

import (
	"context"

	"github.com/werf/werf/pkg/giterminism_inspector"

	"github.com/werf/logboek"
)

var (
	GlobalWarningLines []string
)

func PrintGlobalWarnings(ctx context.Context) {
	for _, line := range GlobalWarningLines {
		printGlobalWarningLn(ctx, line)
	}

	giterminism_inspector.PrintInspectionDebrief(ctx)
}

func GlobalWarningLn(ctx context.Context, line string) {
	GlobalWarningLines = append(GlobalWarningLines, line)
	printGlobalWarningLn(ctx, line)
}

func printGlobalWarningLn(ctx context.Context, line string) {
	logboek.Context(ctx).Error().LogF("WARNING: %s\n", line)
}
