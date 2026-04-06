package validate

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Validate CycloneDX JSON SBOM files against ISPRAS schemas using sbom-checker.

The command runs sbom-checker inside a Docker container and reports validation results. Supports both OSS and container SBOM types.`

	docs.LongMD = "Validate CycloneDX JSON SBOM files against ISPRAS schemas using sbom-checker.\n\n" +
		"The command runs sbom-checker inside a Docker container and reports validation results. " +
		"Supports both OSS and container SBOM types."

	return docs
}
