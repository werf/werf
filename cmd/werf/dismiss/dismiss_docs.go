package dismiss

import (
	"fmt"

	"github.com/werf/werf/cmd/werf/docs/structs"
	"github.com/werf/werf/pkg/werf"
)

func GetDismissDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = fmt.Sprintf(`Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm Release name, Kubernetes Namespace and how to change it: https://%s/documentation/usage/deploy/environments.html`, werf.Domain)

	docs.LongMD = "Delete application from Kubernetes.\n\n" +
		"Helm Release will be purged and optionally Kubernetes Namespace.\n\n" +
		"Environment is a required param for the dismiss by default, because " +
		"it is needed to construct Helm Release name and Kubernetes Namespace. " +
		"Either `--env` or `$WERF_ENV` should be specified for command.\n\n" +
		"Read more info about Helm Release name, Kubernetes Namespace and how " +
		fmt.Sprintf("to change it: https://%s/documentation/usage/deploy/environments.html", werf.Domain)

	return docs
}
