package purge

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetPurgeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Purge all project images in the container registry.

WARNING: Images that are being used in the Kubernetes cluster will also be deleted.`

	docs.LongMD = "Purge all project images in the container registry.\n\n" +
		"**WARNING**: Images that are being used in the Kubernetes cluster will also be deleted."

	return docs
}
