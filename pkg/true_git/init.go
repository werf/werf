package true_git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	MinGitVersionConstraintValue               = "1.9"
	MinGitVersionWithSubmodulesConstraintValue = "2.14"
)

var (
	gitVersion *semver.Version

	minGitVersionErrorMsg     = fmt.Sprintf("Git version >= %s required", MinGitVersionConstraintValue)
	submodulesVersionErrorMsg = fmt.Sprintf("To use git submodules install git >= %s", MinGitVersionWithSubmodulesConstraintValue)

	outStream, errStream io.Writer
)

type Options struct {
	Out, Err io.Writer
}

func Init(opts Options) error {
	outStream = os.Stdout
	errStream = os.Stderr
	if opts.Out != nil {
		outStream = opts.Out
	}
	if opts.Err != nil {
		errStream = opts.Err
	}

	var err error

	v, err := getGitCliVersion()
	if err != nil {
		return err
	}

	vObj, err := semver.NewVersion(v)
	if err != nil {
		errMsg := strings.Join([]string{
			fmt.Sprintf("unexpected git version spec %s", v),
			minGitVersionErrorMsg,
			submodulesVersionErrorMsg,
		}, ".\n")

		return errors.New(errMsg)
	}
	gitVersion = vObj

	if err := checkMinVersionConstraint(); err != nil {
		return err
	}

	return nil
}

func getGitCliVersion() (string, error) {
	cmd := exec.Command("git", "version")

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		errMsg := strings.Join([]string{
			fmt.Sprintf("git version command failed: %s", err),
			minGitVersionErrorMsg,
			submodulesVersionErrorMsg,
		}, ".\n")

		return "", errors.New(errMsg)
	}

	strippedOut := strings.TrimSpace(out.String())
	rightPart := strings.TrimLeft(strippedOut, "git version ")
	fullVersion := strings.Split(rightPart, " ")[0]
	fullVersionParts := strings.Split(fullVersion, ".")

	lowestVersionPartInd := 3
	if len(fullVersionParts) < lowestVersionPartInd {
		lowestVersionPartInd = len(fullVersionParts)
	}

	version := strings.Join(fullVersionParts[0:lowestVersionPartInd], ".")

	return version, nil
}

func checkMinVersionConstraint() error {
	constraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionConstraintValue))
	if err != nil {
		panic(err)
	}

	if !constraint.Check(gitVersion) {
		errMsg := strings.Join([]string{
			strings.ToLower(minGitVersionErrorMsg),
			submodulesVersionErrorMsg,
			fmt.Sprintf("Your git version is %s", gitVersion.String()),
		}, ".\n")

		return errors.New(errMsg)
	}

	return nil
}

func checkSubmoduleConstraint() error {
	constraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionWithSubmodulesConstraintValue))
	if err != nil {
		panic(err)
	}

	if !constraint.Check(gitVersion) {
		errMsg := strings.Join([]string{
			strings.ToLower(submodulesVersionErrorMsg),
			fmt.Sprintf("Your git version is %s", gitVersion.String()),
		}, ".\n")

		return errors.New(errMsg)
	}

	return nil
}
