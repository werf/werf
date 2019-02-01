package helm

import (
	"fmt"
	"strings"

	version "github.com/hashicorp/go-version"
)

func ValidateHelmVersion() error {
	ver, err := HelmVersion()
	if err != nil {
		return err
	}

	lowestVersion, err := version.NewVersion("v2.7.0-rc1")
	if err != nil {
		return err
	}

	if ver.LessThan(lowestVersion) {
		return fmt.Errorf("helm version must be greater than '%s' (current version '%s')", lowestVersion.String(), ver.String())
	}

	return nil
}

func HelmVersion() (*version.Version, error) {
	stdout, stderr, err := HelmCmd("version", "--client", "--short")
	if err != nil {
		return nil, fmt.Errorf("unable to get helm version: %v\n%v %v", err, stdout, stderr)
	}

	parts := strings.Split(stdout, " ")
	vParts := strings.Split(parts[1], "+")
	v := vParts[0]

	ver, err := version.NewVersion(v)
	if err != nil {
		return nil, err
	}

	return ver, nil
}
