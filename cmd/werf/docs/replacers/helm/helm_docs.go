package helm

import "github.com/werf/werf/cmd/werf/docs/structs"

func GetHelmCreateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command creates a chart directory along with the common files and" +
		"directories used in a chart.\n\n" +
		"For example, `helm create foo` will create a directory structure that looks\n" +
		"something like this:\n" +
		"```\n" +
		"foo/\n" +
		"├── .helmignore   # Contains patterns to ignore when packaging Helm charts.\n" +
		"├── Chart.yaml    # Information about your chart.\n" +
		"├── values.yaml   # The default values for your templates.\n" +
		"├── charts/       # Charts that this chart depends on.\n" +
		"└── templates/    # The template files.\n" +
		"    └── tests/    # The test files.\n" +
		"```\n" +
		"`helm create` takes a path for an argument. If directories in the given path\n" +
		"do not exist, Helm will attempt to create them as it goes. If the given\n" +
		"destination exists and there are files in that directory, conflicting files\n" +
		"will be overwritten, but other files will be left alone.\n"

	return docs
}

func GetHelmEnvDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "`env` prints out all the environment information in use by Helm."

	return docs
}
func GetHelmHistoryDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "History prints historical revisions for a given release.\n\n" +
		"A default maximum of 256 revisions will be returned. Setting `--max` " +
		"configures the maximum length of the revision list returned.\n\n" +
		"The historical release set is printed as a formatted table, e.g:\n" +
		"```\n" +
		"$ helm history angry-bird\n" +
		"REVISION    UPDATED                     STATUS          CHART             APP VERSION     DESCRIPTION\n" +
		"1           Mon Oct 3 10:15:13 2016     superseded      alpine-0.1.0      1.0             Initial install\n" +
		"2           Mon Oct 3 10:15:13 2016     superseded      alpine-0.1.0      1.0             Upgraded successfully\n" +
		"3           Mon Oct 3 10:15:13 2016     superseded      alpine-0.1.0      1.0             Rolled back to 2\n" +
		"4           Mon Oct 3 10:15:13 2016     deployed        alpine-0.1.0      1.0             Upgraded successfully\n" +
		"```"

	return docs
}
