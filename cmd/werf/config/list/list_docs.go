package list

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetListDocs() structs.DocsShortStruct {
	var docs structs.DocsShortStruct

	docs.Short = "List image and artifact names defined in werf.yaml."
	docs.ShortMD = "List image and artifact names defined in `werf.yaml`."

	return docs
}
