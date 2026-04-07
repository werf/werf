package merge

import "github.com/werf/werf/v2/cmd/werf/docs/structs"

func GetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Merge per-image CycloneDX 1.6 SBOMs into a single product/module-level SBOM.

Takes a JSON mapping file (image name -> sha256 digest) as input, pulls per-image SBOMs from the container registry, and produces an aggregated SBOM with preserved dependency graphs.

Two ISPRAS-defined output formats are supported:
- "container": hierarchical — each image becomes a top-level container component with nested packages.
- "oss": flat — all packages from all images are merged into a deduplicated flat list.

GOST properties (attack_surface, security_function) are aggregated bottom-up using the "yes > indirect > no" precedence rule.`

	docs.LongMD = "Merge per-image CycloneDX 1.6 SBOMs into a single product/module-level SBOM.\n\n" +
		"Takes a JSON mapping file (image name → sha256 digest) as input, pulls per-image SBOMs " +
		"from the container registry, and produces an aggregated SBOM with preserved dependency graphs.\n\n" +
		"Two ISPRAS-defined output formats are supported:\n" +
		"- `container`: hierarchical — each image becomes a top-level container component with nested packages.\n" +
		"- `oss`: flat — all packages from all images are merged into a deduplicated flat list.\n\n" +
		"GOST properties (`attack_surface`, `security_function`) are aggregated bottom-up using " +
		"the `yes > indirect > no` precedence rule."

	return docs
}
