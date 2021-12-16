package global_warnings

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/Masterminds/semver"

	"github.com/werf/logboek"
)

const LastMultiwerfVersion = "1.5.0"

var GlobalWarningLines []string

func PrintGlobalWarnings(ctx context.Context) {
	for _, line := range GlobalWarningLines {
		printGlobalWarningLn(ctx, line)
	}
}

func GlobalWarningLn(ctx context.Context, line string) {
	GlobalWarningLines = append(GlobalWarningLines, line)
	printGlobalWarningLn(ctx, line)
}

func IsMultiwerfUpToDate() (bool, error) {
	multiwerfPath, err := exec.LookPath("multiwerf")
	if err != nil {
		return true, nil
	}

	cmd := exec.Command(multiwerfPath, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("unable to get installed version of multiwerf: %s", err)
	}

	versionRegex := regexp.MustCompile(`^multiwerf v([.0-9]+)\s*$`)
	regexResult := versionRegex.FindStringSubmatch(stdout.String())
	if len(regexResult) != 2 {
		return false, fmt.Errorf("\"multiwerf version\" returned unexpected output: %s", stdout.String())
	}
	installedMultiwerfVersion, err := semver.NewVersion(regexResult[1])
	if err != nil {
		return false, fmt.Errorf("unable to parse version of installed multiwerf version: %s", err)
	}

	lastMultiwerfVersion, err := semver.NewVersion(LastMultiwerfVersion)
	if err != nil {
		return false, fmt.Errorf("unable to parse version of last available multiwerf version: %s", err)
	}

	return !installedMultiwerfVersion.LessThan(lastMultiwerfVersion), nil
}

func PostponeMultiwerfNotUpToDateWarning() {
	if multiwerfIsUpToDate, err := IsMultiwerfUpToDate(); err != nil {
		GlobalWarningLines = append(
			GlobalWarningLines,
			fmt.Sprintf("Failure detecting whether multiwerf (if present) is outdated: %s", err),
			"Multiwerf is deprecated, so if you are still using it we strongly recommend removing multiwerf and switching to trdl by following these instructions: https://werf.io/installation.html",
		)
		return
	} else if multiwerfIsUpToDate {
		return
	}

	GlobalWarningLines = append(
		GlobalWarningLines,
		"Multiwerf detected, but is out of date. Multiwerf is deprecated in favor of trdl: https://github.com/werf/trdl",
		"If you are still using multiwerf we strongly recommend removing multiwerf and switching to trdl by following these instructions: https://werf.io/installation.html",
	)
}

func printGlobalWarningLn(ctx context.Context, line string) {
	logboek.Context(ctx).Error().LogF("WARNING: %s\n", line)
}
