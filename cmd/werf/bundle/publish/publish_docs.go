package publish

import (
	"github.com/werf/werf/v3/cmd/werf/docs/structs"
)

func GetBundlePublishDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Publish bundle into the container registry. The command builds the images defined in the werf.yaml and publishes the Helm chart together with them.

The chart files are published as they are, with the chart name and version set from the repo and the tag. The publish-time inputs are stored in the .werf-data directory of the chart: the --values, the --secret-values (still encrypted, so no secret key is needed), the --set* params with the contents of the --set-file files, the --patches, the service values with the built images, the annotations, the labels and the release name and namespace. The built images are also listed in the helm.sh/images annotation of the Chart.yaml.

Published into container registry bundle can be rolled out by the "werf bundle apply" command, which applies the stored publish-time inputs the same way "werf converge" applies them.`

	docs.LongMD = "Publish bundle into the container registry. The command builds the images defined in the `werf.yaml` " +
		"and publishes the Helm chart together with them.\n\n" +
		"The chart files are published as they are, with the chart name and version set from the repo and the tag. " +
		"The publish-time inputs are stored in the `.werf-data` directory of the chart: the `--values`, " +
		"the `--secret-values` (still encrypted, so no secret key is needed), the `--set*` params with the contents " +
		"of the `--set-file` files, the `--patches`, the service values with the built images, the annotations, " +
		"the labels and the release name and namespace. The built images are also listed in the `helm.sh/images` annotation " +
		"of the `Chart.yaml`, which Helm-based tools read.\n\n" +
		"Published into container registry bundle can be rolled out by the `werf bundle apply` command, which applies " +
		"the stored publish-time inputs the same way `werf converge` applies them."

	return docs
}
