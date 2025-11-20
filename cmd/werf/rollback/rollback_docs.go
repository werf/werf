package rollback

import (
	"github.com/werf/werf/v2/cmd/werf/docs/structs"
)

func GetRollbackDocs() structs.DocsStruct {
	var docs structs.DocsStruct
	docs.Long = `Rollback the Helm release.

Rollback to the previous revision by default. Use --revision to specify the revision to rollback to. Should be run inside of a werf project directory, otherwise --namespace and --release must be specified.

The result of rollback command is an application deployed into Kubernetes. Command will create the new release based on the one we rollback to and wait until all resources of the release will become ready.`

	docs.LongMD = "Rollback the Helm release.\n\n" +
		"The result of rollback command is an application deployed into Kubernetes. " +
		"Command will create the new release based on the one we rollback to and wait until all resources of the release will become ready.\n\n"
	return docs
}
