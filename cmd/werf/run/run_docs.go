package run

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetRunDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Run container for specified project image from werf.yaml (build if needed).`

	docs.LongMD = "Run container for specified project image from `werf.yaml` (build if needed)."

	return docs
}
