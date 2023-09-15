package plan

import (
	"github.com/werf/werf/cmd/werf/docs/structs"
)

func GetPlanDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = `Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.

Environment is a required param by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/documentation/usage/deploy/environments.html`

	docs.LongMD = "Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy.\n\n" +
		"Environment is a required param by default, because it is needed to construct " +
		"Helm Release name and Kubernetes Namespace. Either `--env` or `$WERF_ENV` should be specified " +
		"for command.\n\n" +
		"Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and " +
		"how to change it: https://werf.io/documentation/usage/deploy/environments.html\n"
	return docs
}
