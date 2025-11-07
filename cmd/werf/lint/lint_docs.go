package lint

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetLintDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Lint Helm chart. This command will calculate digests and build (if needed) all images defined in the werf.yaml.`

	docs.LongMD = "Lint Helm chart. This command will calculate digests and build " +
		"(if needed) all images defined in the `werf.yaml`."

	return docs
}
