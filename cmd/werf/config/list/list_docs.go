package list

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetListDocs() structs.DocsShortStruct {
	var docs structs.DocsShortStruct

	docs.Short = "List image names defined in werf.yaml."
	docs.ShortMD = "List image names defined in `werf.yaml`."

	return docs
}
