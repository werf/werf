package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	MinGitVersionConstraint               = "1.9.1"
	MinGitVersionWithSubmodulesConstraint = "2.14.0"
)

var (
	GitVersion string

	gitVersionObj                 *semver.Version
	minVersionConstraintObj       *semver.Constraints
	submoduleVersionConstraintObj *semver.Constraints
)

func Init() error {
	var err error

	v, vObj, err := getGitCliVersion()
	if err != nil {
		return err
	}
	GitVersion = v
	gitVersionObj = vObj

	var c *semver.Constraints

	c, err = semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionConstraint))
	if err != nil {
		panic(err)
	}
	minVersionConstraintObj = c

	if !minVersionConstraintObj.Check(gitVersionObj) {
		return fmt.Errorf("Git version >= %s required! To use submodules install git >= %s. Your git version is %s.", MinGitVersionConstraint, MinGitVersionWithSubmodulesConstraint, GitVersion)
	}

	c, err = semver.NewConstraint(fmt.Sprintf(">= %s", MinGitVersionWithSubmodulesConstraint))
	if err != nil {
		panic(err)
	}
	submoduleVersionConstraintObj = c

	return nil
}

func getGitCliVersion() (string, *semver.Version, error) {
	cmd := exec.Command("git", "--version")

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", nil, fmt.Errorf("git command is not available")
	}

	versionParts := strings.Split(out.String(), " ")
	if len(versionParts) != 3 {
		return "", nil, fmt.Errorf("unexpected `git --version` output `%s`", out.String())
	}
	rawVersion := strings.TrimSpace(versionParts[2])

	version, err := semver.NewVersion(rawVersion)
	if err != nil {
		return "", nil, fmt.Errorf("unexpected version spec `%s`: %s", rawVersion, err)
	}

	return rawVersion, version, nil
}
