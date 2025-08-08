package global_warnings

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
)

const lastMultiwerfVersion = "1.5.0"

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

func IsMultiwerfUpToDate(ctx context.Context) (bool, error) {
	multiwerfPath, err := exec.LookPath("multiwerf")
	if err != nil {
		return true, nil
	}

	cmd := werfExec.CommandContextCancellation(ctx, multiwerfPath, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		werfExec.TerminateIfCanceled(ctx, context.Cause(ctx), werfExec.ExitCode(err))
		return false, fmt.Errorf("unable to get installed version of multiwerf: %w", err)
	}

	versionRegex := regexp.MustCompile(`^multiwerf v([.0-9]+)\s*$`)
	regexResult := versionRegex.FindStringSubmatch(stdout.String())
	if len(regexResult) != 2 {
		return false, fmt.Errorf("\"multiwerf version\" returned unexpected output: %s", stdout.String())
	}
	installedMultiwerfVersion, err := semver.NewVersion(regexResult[1])
	if err != nil {
		return false, fmt.Errorf("unable to parse version of installed multiwerf version: %w", err)
	}

	lastMultiwerfVersion, err := semver.NewVersion(lastMultiwerfVersion)
	if err != nil {
		return false, fmt.Errorf("unable to parse version of last available multiwerf version: %w", err)
	}

	return !installedMultiwerfVersion.LessThan(lastMultiwerfVersion), nil
}

func PostponeMultiwerfNotUpToDateWarning(ctx context.Context) {
	if multiwerfIsUpToDate, err := IsMultiwerfUpToDate(ctx); err != nil {
		msg := fmt.Sprintf(`Failure detecting whether multiwerf (if present) is outdated: %s
multiwerf is deprecated, so if you are still using it we strongly recommend removing multiwerf and switching to trdl
`, err)

		globalWarningMessages = append(
			globalWarningMessages,
			msg,
		)
		return
	} else if multiwerfIsUpToDate {
		return
	}

	globalWarningMessages = append(
		globalWarningMessages,
		`multiwerf detected, but it is out of date. multiwerf is deprecated in favor of trdl: https://github.com/werf/trdl
If you are still using multiwerf we strongly recommend removing multiwerf and switching to trdl`,
	)
}

func printGlobalWarningLn(ctx context.Context, line string) {
	if SuppressGlobalWarnings {
		return
	}
	logboek.Context(ctx).Warn().LogLn(line)
}
