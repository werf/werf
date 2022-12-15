package converge

import (
	"github.com/werf/werf/cmd/werf/docs/structs"
)

func GetConvergeDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = `Build and push images, then deploy application into Kubernetes.

The result of converge command is an application deployed into Kubernetes for current git state. Command will create release and wait until all resources of the release will become ready.

Environment is a required param for the deploy by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/documentation/usage/deploy/environments.html`

	docs.LongMD = "Build and push images, then deploy application into Kubernetes.\n\n" +
		"The result of converge command is an application deployed into Kubernetes for current git state. " +
		"Command will create release and wait until all resources of the release will become ready.\n\n" +
		"Environment is a required param for the deploy by default, because it is needed to construct " +
		"Helm Release name and Kubernetes Namespace. Either `--env` or `$WERF_ENV` should be specified " +
		"for command.\n\n" +
		"Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and " +
		"how to change it: https://werf.io/documentation/usage/deploy/environments.html\n"
	return docs
}
