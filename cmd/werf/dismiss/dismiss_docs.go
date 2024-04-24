package dismiss

import (
	"github.com/werf/werf/v2/cmd/werf/docs/structs"
)

func GetDismissDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.`

	docs.LongMD = "Delete application from Kubernetes.\n\n" +
		"Helm Release will be purged and optionally Kubernetes Namespace.\n\n" +
		"Environment is a required param for the dismiss by default, because " +
		"it is needed to construct Helm Release name and Kubernetes Namespace. " +
		"Either `--env` or `$WERF_ENV` should be specified for command.\n\n"

	return docs
}
