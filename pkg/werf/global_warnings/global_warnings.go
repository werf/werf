package global_warnings

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
)

var (
	globalWarningMessages            []string
	globalDeprecationWarningMessages []string
	SuppressGlobalWarnings           bool
)

func PrintGlobalWarnings(ctx context.Context) {
	printFunc := func(header string, messages []string) {
		if len(messages) == 0 {
			return
		}

		printGlobalWarningLn(ctx, header)
		for i, line := range util.UniqStrings(messages) {
			if i != 0 {
				printGlobalWarningLn(ctx, "")
			}

			multilineLines := strings.Split(line, "\n")
			for j, mlLine := range multilineLines {
				if j == 0 {
					printGlobalWarningLn(ctx, fmt.Sprintf("%d: %s", i+1, mlLine))
				} else {
					printGlobalWarningLn(ctx, fmt.Sprintf("   %s", mlLine))
				}
			}
		}
	}

	printFunc("DEPRECATION WARNINGS:", globalDeprecationWarningMessages)
	if len(globalWarningMessages) > 0 {
		printGlobalWarningLn(ctx, "")
		printFunc("WARNINGS:", globalWarningMessages)
	}
}

func GlobalDeprecationWarningLn(ctx context.Context, line string) {
	globalDeprecationWarningMessages = append(globalDeprecationWarningMessages, line)
	printGlobalWarningLn(ctx, "DEPRECATION WARNING! "+line)
}

func GlobalWarningLn(ctx context.Context, line string) {
	globalWarningMessages = append(globalWarningMessages, line)
	printGlobalWarningLn(ctx, "WARNING! "+line)
}

func printGlobalWarningLn(ctx context.Context, line string) {
	if SuppressGlobalWarnings {
		return
	}
	logboek.Context(ctx).Warn().LogLn(line)
}
