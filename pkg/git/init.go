package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	MinGitVersionConstraint               = "1.9.0"
	MinGitVersionWithSubmodulesConstraint = "2.14.0"
)

var (
	GitVersion            string
	RequiredGitVersionMsg = fmt.Sprintf("Git version >= %s required!", MinGitVersionConstraint)
	// Uncomment when submodules supported by go-dapp
	// RequiredGitVersionMsg = fmt.Sprintf("Git version >= %s required! To use submodules install git >= %s.", MinGitVersionConstraint, MinGitVersionWithSubmodulesConstraint)

	gitVersionObj                 *semver.Version
	minVersionConstraintObj       *semver.Constraints
	submoduleVersionConstraintObj *semver.Constraints
)

func Init() error {
	var err error

	v, err := getGitCliVersion()
	if err != nil {
		return err
	}
	GitVersion = v

	vObj, err := semver.NewVersion(GitVersion)
	if err != nil {
		return fmt.Errorf("unexpected `git --version` spec `%s`: %s\n%s\nYour git version is %s.", GitVersion, err, RequiredGitVersionMsg, GitVersion)
	}
	gitVersionObj = vObj

	var c *semver.Constraints

	c, err = semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionConstraint))
	if err != nil {
		panic(err)
	}
	minVersionConstraintObj = c

	if !minVersionConstraintObj.Check(gitVersionObj) {
		return fmt.Errorf("%s\nYour git version is %s.", RequiredGitVersionMsg, GitVersion)
	}

	c, err = semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionWithSubmodulesConstraint))
	if err != nil {
		panic(err)
	}
	submoduleVersionConstraintObj = c

	return nil
}

func getGitCliVersion() (string, error) {
	cmd := exec.Command("git", "--version")

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command is not available!\n%s", RequiredGitVersionMsg)
	}

	versionParts := strings.Split(out.String(), " ")
	if len(versionParts) != 3 {
		return "", fmt.Errorf("unexpected `git --version` output:\n\n```\n%s\n```\n\n%s", strings.TrimSpace(out.String()), RequiredGitVersionMsg)
	}
	rawVersion := strings.TrimSpace(versionParts[2])

	return rawVersion, nil
}

func checkSubmoduleConstraint() error {
	if !submoduleVersionConstraintObj.Check(gitVersionObj) {
		return fmt.Errorf("To use submodules install git >= %s! Your git version is %s.", MinGitVersionWithSubmodulesConstraint, GitVersion)
	}
	return nil
}
