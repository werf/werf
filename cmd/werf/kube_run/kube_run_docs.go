package kube_run

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetKubeRunDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Run container in Kubernetes for specified project image from werf.yaml (build if needed).`

	docs.LongMD = "Run container in Kubernetes for specified project image from `werf.yaml` (build if needed)."

	return docs
}
