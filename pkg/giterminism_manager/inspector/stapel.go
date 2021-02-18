package inspector

import (
	"fmt"
)

func (i Inspector) InspectConfigStapelFromLatest() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelFromLatestAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError(`fromLatest directive not allowed by giterminism

If fromLatest is true, then werf starts using the actual base image digest in the stage digest. Thus, using this directive may break the reproducibility of previous builds. The changing of the base image in the registry makes all previously built images unusable.

 * Previous pipeline jobs (e.g., converge) cannot be retried without the image rebuilding after changing a registry base image.
 * If the base image is modified unexpectedly, it may lead to an inexplicably failed pipeline. For instance, the modification occurs after a successful build, and the following jobs will be failed due to changing stages digests alongside base image digest.

As an alternative, we recommend using unchangeable tag or periodically change 'fromCacheVersion' value to guarantee the application's controllable and predictable life cycle.`)
}

func (i Inspector) InspectConfigStapelGitBranch() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelGitBranchAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError(`git branch directive not allowed by giterminism

Remote git mapping with a branch (master branch by default) may break the previous builds' reproducibility. werf uses the history of a git repository to calculate the stage digest. Thus, the new commit in the branch makes all previously built images unusable.

 * The existing pipeline jobs (e.g., converge) would not run and would require rebuilding an image if a remote git branch has been changed.
 * Unplanned commits to a remote git branch might lead to the pipeline failing seemingly for no apparent reasons. For instance, changes may occur after the build process is completed successfully. In this case, the related pipeline jobs will fail due to changes in stage digests along with the branch HEAD.

As an alternative, we recommend using unchangeable reference, tag, or commit to guarantee the application's controllable and predictable life cycle.`)
}

func (i Inspector) InspectConfigStapelMountBuildDir() error {
	if i.sharedOptions.LooseGiterminism() || i.giterminismConfig.IsConfigStapelMountBuildDirAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError(`"mount { from: build_dir, ... }" not allowed by giterminism

The use of the build_dir mount may lead to unpredictable behavior when used in parallel and potentially affect reproducibility and reliability.`)
}

func (i Inspector) InspectConfigStapelMountFromPath(fromPath string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigStapelMountFromPathAccepted(fromPath) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(`"mount { fromPath: %s, ... }" not allowed by giterminism

The use of the fromPath mount may lead to unpredictable behavior when used in parallel and potentially affect reproducibility and reliability. The data in the mounted directory has no effect on the final image digest, which can lead to invalid images and hard-to-trace issues.`, fromPath))
}
