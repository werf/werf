package reset

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetPurgeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Purge werf images, cache and other data for all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, converge and other commands.
* Local cache:
  * remote Git clones cache;
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually.`

	docs.LongMD = "Purge werf images, cache and other data for all projects on host machine.\n\n" +
		"The data include:\n" +
		"* Old service tmp dirs, which werf creates during every build, converge and other commands.\n" +
		"* Local cache:\n" +
		"  * remote Git clones cache;\n" +
		"  * Git worktree cache.\n" +
		"* Shared context:\n" +
		"  * Mounts which persists between several builds (mounts from build_dir).\n\n" +
		"**WARNING**: Do not run this command during any other werf command is working on " +
		"the host machine. This command is supposed to be run manually."

	return docs
}
