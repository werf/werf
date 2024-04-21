package plan

import (
	"github.com/werf/werf/v2/cmd/werf/docs/structs"
)

func GetPlanDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = `Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.

Environment is a required param by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.`

	docs.LongMD = "Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.\n\n" +
		"Environment is a required param by default, because it is needed to construct " +
		"Helm Release name and Kubernetes Namespace. Either `--env` or `$WERF_ENV` should be specified " +
		"for command.\n\n"
	return docs
}
