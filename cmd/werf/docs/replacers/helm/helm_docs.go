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
