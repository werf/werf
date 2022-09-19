package publish

import "github.com/werf/werf/cmd/werf/common/docs_struct_templates"

func GetBundlePublishDocs() docs_struct_templates.DocsStruct {
	var docs docs_struct_templates.DocsStruct

	docs.Long = `Publish bundle into the container registry. werf bundle contains built images defined in the werf.yaml, Helm chart, Service values which contain built images tags, any custom values and set values params provided during publish invocation, werf addon templates (like werf_image).

Published into container registry bundle can be rolled out by the "werf bundle" command.`

	docs.LongMD = "Publish bundle into the container registry. `werf bundle` contains built images defined " +
		"in the `werf.yaml`, Helm chart, Service values which contain built images tags, any custom values " +
		"and set values params provided during publish invocation, werf addon templates (like `werf_image`).\n\n" +
		"Published into container registry bundle can be rolled out by the `werf bundle` command."

	return docs
}
