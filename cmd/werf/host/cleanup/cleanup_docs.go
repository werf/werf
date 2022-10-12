package cleanup

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetCleanupDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Cleanup old unused werf cache and data of all projects on host machine.

The data include:
* Lost Docker containers and images from interrupted builds.
* Old service tmp dirs, which werf creates during every build, converge and other commands.
* Local cache:
  * remote Git clones cache;
  * Git worktree cache.

It is safe to run this command periodically by automated cleanup job in parallel with other werf commands such as build, converge and cleanup.`

	docs.LongMD = "Cleanup old unused werf cache and data of all projects on host machine.\n\n" +
		"The data include:\n" +
		"* Lost Docker containers and images from interrupted builds.\n" +
		"* Old service tmp dirs, which werf creates during every `build`, `converge` and other commands.\n" +
		"* Local cache:\n" +
		"  * remote Git clones cache;\n" +
		"  * Git worktree cache.\n\n" +
		"It is safe to run this command periodically by automated cleanup job in parallel with " +
		"other werf commands such as `build`, `converge` and `cleanup`."

	return docs
}
