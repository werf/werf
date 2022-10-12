package render

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetRenderDocs() structs.DocsShortStruct {
	var docs structs.DocsShortStruct

	docs.Short = "Render werf.yaml."
	docs.ShortMD = "Render `werf.yaml`."

	return docs
}
