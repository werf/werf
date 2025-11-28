package copy

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetCopyDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = `Copy project stages between storages.

Supported copy scenarios:
	• Container registry to container registry
	• Container registry to archive
	• Archive to container registry

By default copies all project stages. Use --all=false to copy only stages relevant for the current git commit. This requires a working directory with git repository and performs build to identify current stages.

Note: this flag is ignored when copying from archive to container registry.
`

	docs.LongMD = "Copy project stages between container registry and archive storage.\n\n" +
		"Supported copy scenarios:\n\n" +
		"\t• Container registry to container registry\n\n" +
		"\t• Container registry to archive\n\n" +
		"\t• Archive to container registry\n\n" +
		"By default copies all project stages. Use --all=false to copy only stages relevant for the current git commit. This requires a working directory with git repository and performs build to identify current stages.\n\n" +
		"Note: this flag is ignored when copying from archive to container registry.\n\n"

	return docs
}
