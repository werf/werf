package inspector

import (
	"fmt"
)

func (i Inspector) InspectConfigStapelFromLatest() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelFromLatestAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError(`fromLatest directive not allowed

Pay attention, werf uses actual base image digest in stage digest if 'fromLatest' is specified. Thus, the usage of this directive might break the reproducibility of previous builds. If the base image is changed in the registry, all previously built stages become not usable.

 * Previous pipeline jobs (e.g. deploy) cannot be retried without the image rebuild after changing base image in the registry.
 * If base image is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following jobs will be failed due to changing of stages digests alongside base image digest.

We recommend a particular unchangeable tag or periodically change 'fromCacheVersion' value to provide controllable and predictable lifecycle of software.`)
}

func (i Inspector) InspectConfigStapelGitBranch() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelGitBranchAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError("git branch directive not allowed")
}

func (i Inspector) InspectConfigStapelMountBuildDir() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelMountBuildDirAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError("'mount { from: build_dir, ... }' not allowed")
}

func (i Inspector) InspectConfigStapelMountFromPath(fromPath string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.giterminismConfig.IsConfigStapelMountFromPathAccepted(fromPath); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf("'mount { fromPath: %s, ... }' not allowed", fromPath))
}
