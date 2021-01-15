package inspector

import (
	"fmt"
)

func (i Inspector) InspectConfigStapelGitBranch() error {
	if i.manager.LooseGiterminism() || i.manager.Config().IsConfigStapelGitBranchAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError("git branch directive not allowed")
}

func (i Inspector) InspectConfigStapelMountBuildDir() error {
	if i.manager.LooseGiterminism() || i.manager.Config().IsConfigStapelMountBuildDirAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError("'mount { from: build_dir, ... }' not allowed")
}

func (i Inspector) InspectConfigStapelMountFromPath(fromPath string) error {
	if i.manager.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.manager.Config().IsConfigStapelMountFromPathAccepted(fromPath); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf("'mount { fromPath: %s, ... }' not allowed", fromPath))
}
