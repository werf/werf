package verify

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetVerifyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Use = "verify --image-ref <image-ref> [--image-ref <image-ref>...] --verify-roots=<root-ca.pem> [--verify-manifest] [--verify-elf-files]"

	docs.Short = "Verify images"

	docs.Long = `Verify the integrity of built images by checking signatures of manifests and ELF files using specific image references.

The command outputs detailed logs about the verification process.`

	docs.LongMD = "Verify the integrity of built images by checking signatures of manifests and ELF files using specific image references.\n\n" +
		"The command outputs detailed logs about the verification process."

	docs.Example = `  # Verify image manifest and ELF files for specific reference from Docker Hub
  $ werf verify --image-ref <DOCKER HUB USERNAME>/werf-guide-app:f4caaa836701e5346c4a0514bb977362ba5fe4ae114d0176f6a6c8cc-1612277803607 --verify-roots=/tmp/root-ca.pem --verify-manifest --verify-elf-files`

	return docs
}
