package build

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetBuildDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Build images that are described in werf.yaml.

The result of build command is built images pushed into the specified repo (or locally if repo is not specified).

If one or more IMAGE_NAME parameters specified, werf will build only these images.`

	docs.LongMD = "Build images that are described in `werf.yaml`.\n\n" +
		"The result of `build` command is built images pushed into the specified " +
		"repo (or locally if repo is not specified).\n\n" +
		"If one or more `IMAGE_NAME` parameters specified, werf will build only these images."

	return docs
}
