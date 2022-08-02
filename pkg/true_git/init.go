package true_git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	MinGitVersionConstraintValue = "2.18"
)

var ForbiddenGitVersionsConstraintValues = []string{"2.22.0"}

var (
	gitVersion *semver.Version

	minGitVersionErrorMsg       = fmt.Sprintf("Git version >= %s required", MinGitVersionConstraintValue)
	forbiddenGitVersionErrorMsg = fmt.Sprintf("Forbidden git versions: %s", strings.Join(ForbiddenGitVersionsConstraintValues, ", "))

	liveGitOutput bool
)

type Options struct {
	LiveGitOutput bool
}

func Init(ctx context.Context, opts Options) error {
	liveGitOutput = opts.LiveGitOutput

	var err error

	v, err := getGitCliVersion(ctx)
	if err != nil {
		return err
	}

	vObj, err := semver.NewVersion(v)
	if err != nil {
		errMsg := strings.Join([]string{
			fmt.Sprintf("Unexpected git version spec %s", v),
			minGitVersionErrorMsg,
			forbiddenGitVersionErrorMsg,
		}, ".\n")

		return errors.New(errMsg)
	}
	gitVersion = vObj

	if err := checkVersionConstraints(); err != nil {
		return err
	}

	return nil
}

func getGitCliVersion(ctx context.Context) (string, error) {
	versionCmd := NewGitCmd(ctx, nil, "version")
	if err := versionCmd.Run(ctx); err != nil {
		errMsg := strings.Join([]string{
			fmt.Sprintf("git version command failed: %s", err),
			minGitVersionErrorMsg,
			forbiddenGitVersionErrorMsg,
		}, ".\n")

		return "", errors.New(errMsg)
	}

	fullVersionMatch := regexp.MustCompile(`git version ([.0-9]+)`).FindStringSubmatch(versionCmd.OutBuf.String())
	if len(fullVersionMatch) < 2 {
		return "", fmt.Errorf("unable to parse git version from stdout: %s", versionCmd.OutBuf)
	}
	fullVersionParts := strings.Split(fullVersionMatch[1], ".")

	lowestVersionPartInd := 3
	if len(fullVersionParts) < lowestVersionPartInd {
		lowestVersionPartInd = len(fullVersionParts)
	}

	version := strings.Join(fullVersionParts[0:lowestVersionPartInd], ".")

	return version, nil
}

func checkVersionConstraints() error {
	constraints := []*semver.Constraints{}

	if os.Getenv("WERF_DISABLE_GIT_MIN_VERSION_CONSTRAINT") != "1" {
		minVersionConstraints, err := semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionConstraintValue))
		if err != nil {
			panic(err)
		}
		constraints = append(constraints, minVersionConstraints)
	}

	for _, cv := range ForbiddenGitVersionsConstraintValues {
		forbiddenVersionsConstraints, err := semver.NewConstraint(fmt.Sprintf("!= %s", cv))
		if err != nil {
			panic(err)
		}
		constraints = append(constraints, forbiddenVersionsConstraints)
	}

	for i := range constraints {
		if !constraints[i].Check(gitVersion) {
			errMsg := strings.Join([]string{
				minGitVersionErrorMsg,
				forbiddenGitVersionErrorMsg,
				fmt.Sprintf("Your git version is %s", gitVersion.String()),
			}, ".\n")

			return errors.New(errMsg)
		}
	}

	return nil
}
