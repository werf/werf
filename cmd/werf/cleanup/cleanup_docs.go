package cleanup

import (
	"fmt"

	"github.com/werf/werf/cmd/werf/docs/structs"
	"github.com/werf/werf/pkg/werf"
)

func GetCleanupDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = fmt.Sprintf(`Safely cleanup unused project images in the container registry.

The command works according to special rules called cleanup policies, which the user defines in werf.yaml (https://%s/documentation/reference/werf_yaml.html#configuring-cleanup-policies).

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel with other werf commands such as build, converge and host cleanup.`, werf.Domain)

	docs.LongMD = "Safely cleanup unused project images in the container registry.\n\n" +
		"The command works according to special rules called cleanup policies, which the user " +
		fmt.Sprintf("defines in `werf.yaml` (https://%s/documentation/reference/werf_yaml.html#configuring-cleanup-policies).\n\n", werf.Domain) +
		"It is safe to run this command periodically (daily is enough) by automated cleanup job " +
		"in parallel with other werf commands such as `build`, `converge` and `host cleanup`."

	return docs
}
