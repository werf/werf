package plan

import (
	"fmt"

	"github.com/werf/werf/cmd/werf/docs/structs"
	"github.com/werf/werf/pkg/werf"
)

func GetPlanDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = fmt.Sprintf(`Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.

Environment is a required param by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://%s/documentation/usage/deploy/environments.html`, werf.Domain)

	docs.LongMD = "Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.\n\n" +
		"Environment is a required param by default, because it is needed to construct " +
		"Helm Release name and Kubernetes Namespace. Either `--env` or `$WERF_ENV` should be specified " +
		"for command.\n\n" +
		"Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and " +
		fmt.Sprintf("how to change it: https://%s/documentation/usage/deploy/environments.html\n", werf.Domain)
	return docs
}
