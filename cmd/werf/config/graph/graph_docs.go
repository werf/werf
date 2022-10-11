package graph

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetGraphDocs() structs.DocsShortStruct {
	var docs structs.DocsShortStruct

	docs.Short = "Print dependency graph for images in werf.yaml."
	docs.ShortMD = "Print dependency graph for images in `werf.yaml`."

	return docs
}
