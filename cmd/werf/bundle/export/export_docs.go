package export

import (
	"github.com/werf/werf/cmd/werf/docs/structs"
)

func GetBundleExportDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Export bundle into the provided directory (or into directory named as a resulting chart in the current working directory). werf bundle contains built images defined in the werf.yaml, Helm chart, Service values which contain built images tags, any custom values and set values params provided during publish invocation, werf service templates and values.`

	docs.LongMD = "Export bundle into the provided directory (or into directory named as a resulting chart " +
		"in the current working directory). `werf bundle` contains built images defined in the `werf.yaml`, " +
		"Helm chart, Service values which contain built images tags, any custom values and set values params " +
		"provided during publish invocation, werf service templates and values."

	return docs
}
