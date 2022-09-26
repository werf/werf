package export

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetExportDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Export images to an arbitrary repository according to a template specified by the --tag option (build if needed).
The tag may contain the following shortcuts:
- %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there is more than one image in the werf config);
- %image_content_based_tag% to use a content-based tag.
All meta-information related to werf is removed from the exported images, and then images are completely under the user's responsibility.`

	docs.LongMD = "Export images to an arbitrary repository according to a template specified " +
		"by the `--tag` option (build if needed).\n\n" +
		"The tag may contain the following shortcuts:\n" +
		"- `image`, `image_slug` or `image_safe_slug` to use the image name (necessary " +
		"if there is more than one image in the werf config);\n" +
		"- `image_content_based_tag` to use a content-based tag.\n" +
		"All meta-information related to werf is removed from the exported images, and then images " +
		"are completely under the user's responsibility."

	return docs
}
