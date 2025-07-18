package get

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Get SBOM of an image is described in werf.yaml and prints the result to stdout.

If SBOM is not found (locally or remotely) it triggers werf build command to generate SBOM.`

	docs.LongMD = "Get SBOM of an image is described in `werf.yaml and prints the result to stdout`.\n\n" +
		"If SBOM is not found (locally or remotely) it triggers werf build command to generate SBOM."

	return docs
}
