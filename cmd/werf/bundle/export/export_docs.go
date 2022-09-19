package export

import "github.com/werf/werf/cmd/werf/common/docs_struct_templates"

func GetBundleExportDocs() docs_struct_templates.DocsStruct {
	var docs docs_struct_templates.DocsStruct

	docs.Long = `Export bundle into the provided directory (or into directory named as a resulting chart in the current working directory). werf bundle contains built images defined in the werf.yaml, Helm chart, Service values which contain built images tags, any custom values and set values params provided during publish invocation, werf service templates and values.`

	docs.LongMD = "Export bundle into the provided directory (or into directory named as a resulting chart " +
		"in the current working directory). `werf bundle` contains built images defined in the `werf.yaml`, " +
		"Helm chart, Service values which contain built images tags, any custom values and set values params " +
		"provided during publish invocation, werf service templates and values."

	return docs
}
