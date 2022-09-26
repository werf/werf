package render

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetRenderDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Render Kubernetes templates. This command will calculate digests and build (if needed) all images defined in the werf.yaml.`

	docs.LongMD = "Render Kubernetes templates. This command will calculate digests and build " +
		"(if needed) all images defined in the `werf.yaml`."

	return docs
}
