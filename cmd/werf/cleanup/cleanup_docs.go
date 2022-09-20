package cleanup

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetCleanupDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Safely cleanup unused project images in the container registry.

The command works according to special rules called cleanup policies, which the user defines in werf.yaml (https://werf.io/documentation/reference/werf_yaml.html#configuring-cleanup-policies).

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel with other werf commands such as build, converge and host cleanup.`

	docs.LongMD = "Safely cleanup unused project images in the container registry.\n\n" +
		"The command works according to special rules called cleanup policies, which the user " +
		"defines in `werf.yaml` (https://werf.io/documentation/reference/werf_yaml.html#configuring-cleanup-policies).\n\n" +
		"It is safe to run this command periodically (daily is enough) by automated cleanup job " +
		"in parallel with other werf commands such as `build`, `converge` and `host cleanup`."

	return docs
}
