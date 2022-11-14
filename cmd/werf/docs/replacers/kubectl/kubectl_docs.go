package kubectl

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetKubectlDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "kubectl controls the Kubernetes cluster manager.\n\n" +
		"Find more information at: https://kubernetes.io/docs/reference/kubectl/overview/"

	return docs
}

func GetAlphaEventsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Experimental: Display events.\n\n" +
		"Prints a table of the most important information about events. " +
		"You can request events for a namespace, for all namespace, or " +
		"filtered to only those pertaining to a specified resource."

	return docs
}

func GetAlphaDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "These commands correspond to alpha features that are not enabled in Kubernetes clusters by default."

	return docs
}

func GetAnnotateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the annotations on one or more resources.\n\n" +
		"All Kubernetes objects support the ability to store additional data with the object as " +
		"annotations. Annotations are key/value pairs that can be larger than labels and include " +
		"arbitrary string values such as structured JSON. Tools and system extensions may use " +
		"annotations to store their own data.\n\n" +
		"Attempting to set an annotation that already exists will fail unless `--overwrite` is set. " +
		"If `--resource-version` is specified and does not match the current resource version on " +
		"the server the command will fail.\n\n" +
		"Use `kubectl api-resources` for a complete list of supported resources."

	return docs
}

func GetApiResourcesDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Print the supported API resources on the server."

	return docs
}

func GetApiVersionsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Print the supported API versions on the server, in the form of `group/version`."

	return docs
}
