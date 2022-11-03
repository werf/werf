package kubectl

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetAlphaEventsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Experimental: Display events.\n\n" +
		"Prints a table of the most important information about events. " +
		"You can request events for a namespace, for all namespace, or " +
		"filtered to only those pertaining to a specified resource."

	return docs
}
