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

func GetHelmInstallDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command installs a chart archive.\n\n" +
		"The install argument must be a chart reference, a path to a packaged chart, " +
		"a path to an unpacked chart directory or a URL.\n\n" +
		"To override values in a chart, use either the `--values` flag and pass in a file " +
		"or use the `--set` flag and pass configuration from the command line, to force " +
		"a string value use `--set-string`. You can use `--set-file` to set individual " +
		"values from a file when the value itself is too long for the command line " +
		"or is dynamically generated.\n" +
		"```\n$ helm install -f myvalues.yaml myredis ./redis\n```\n" +
		"or\n" +
		"```\n$ helm install --set name=prod myredis ./redis\n```\n" +
		"or\n" +
		"```\n$ helm install --set-string long_int=1234567890 myredis ./redis\n```\n" +
		"or\n" +
		"```\n$ helm install --set-file my_script=dothings.sh myredis ./redis\n```\n" +
		"You can specify the `--values`/`-f` flag multiple times. The priority will be given to the " +
		"last (right-most) file specified. For example, if both `myvalues.yaml` and `override.yaml` " +
		"contained a key called `Test`, the value set in `override.yaml` would take precedence:\n" +
		"```\n$ helm install -f myvalues.yaml -f override.yaml  myredis ./redis\n```\n" +
		"You can specify the `--set` flag multiple times. The priority will be given to the " +
		"last (right-most) set specified. For example, if both `bar` and `newbar` values are " +
		"set for a key called `foo`, the `newbar` value would take precedence:\n" +
		"```\n$ helm install --set foo=bar --set foo=newbar  myredis ./redis\n```\n" +
		"To check the generated manifests of a release without installing the chart, " +
		"the `--debug` and `--dry-run` flags can be combined.\n\n" +
		"If `--verify` is set, the chart **must** have a provenance file, and the provenance " +
		"file **must** pass all verification steps.\n\n" +
		"There are five different ways you can express the chart you want to install:\n" +
		"1. By chart reference: `helm install mymaria example/mariadb`.\n" +
		"2. By path to a packaged chart: `helm install mynginx ./nginx-1.2.3.tgz`.\n" +
		"3. By path to an unpacked chart directory: `helm install mynginx ./nginx`.\n" +
		"4. By absolute URL: `helm install mynginx https://example.com/charts/nginx-1.2.3.tgz`.\n" +
		"5. By chart reference and repo URL: `helm install --repo https://example.com/charts/ mynginx nginx`.\n\n" +
		"### Chart references\n" +
		"A chart reference is a convenient way of referencing a chart in a chart repository.\n\n" +
		"When you use a chart reference with a repo prefix (`example/mariadb`), Helm will look in the local " +
		"configuration for a chart repository named 'example', and will then look for a" +
		"chart in that repository whose name is `mariadb`. It will install the latest stable version of that chart " +
		"until you specify `--devel` flag to also include development version (`alpha`, `beta`, and `release candidate` releases), or " +
		"supply a version number with the `--version` flag.\n\n" +
		"To see the list of chart repositories, use `helm repo list`. To search for " +
		"charts in a repository, use `helm search`."

	return docs
}

func GetHelmLintDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command takes a path to a chart and runs a series of tests to verify that" +
		"the chart is well-formed.\n\n" +
		"If the linter encounters things that will cause the chart to fail installation, " +
		"it will emit `[ERROR]` messages. If it encounters issues that break with convention " +
		"or recommendation, it will emit `[WARNING]` messages."

	return docs
}

func GetHelmListDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command lists all of the releases for a specified namespace " +
		"(uses current namespace context if namespace not specified).\n\n" +
		"By default, it lists only releases that are deployed or failed. Flags like " +
		"`--uninstalled` and `--all` will alter this behavior. Such flags can be combined: " +
		"`--uninstalled --failed`.\n\n" +
		"By default, items are sorted alphabetically. Use the `-d` flag to sort by " +
		"release date.\n\n" +
		"If the `--filter` flag is provided, it will be treated as a filter. Filters are " +
		"regular expressions (Perl compatible) that are applied to the list of releases. " +
		"Only items that match the filter will be returned.\n" +
		"```\n" +
		"    $ helm list --filter 'ara[a-z]+'\n" +
		"    NAME                UPDATED                                  CHART\n" +
		"    maudlin-arachnid    2020-06-18 14:17:46.125134977 +0000 UTC  alpine-0.1.0\n```\n" +
		"If no results are found, `helm list` will exit `0`, but with no output (or in " +
		"the case of no `-q` flag, only headers).\n\n" +
		"By default, up to 256 items may be returned. To limit this, use the `--max` flag. " +
		"Setting `--max` to `0` will not return all results. Rather, it will return the " +
		"server's default, which may be much higher than 256. Pairing the `--max` " +
		"flag with the `--offset` flag allows you to page through results.\n"

	return docs
}

func GetHelmPackageDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command packages a chart into a versioned chart archive file. If a path " +
		"is given, this will look at that path for a chart (which must contain a " +
		"`Chart.yaml` file) and then package that directory.\n\n" +
		"Versioned chart archives are used by Helm package repositories.\n\n" +
		"To sign a chart, use the `--sign` flag. In most cases, you should also " +
		"provide `--keyring path/to/secret/keys` and `--key keyname`.\n" +
		"```\n$ helm package --sign ./mychart --key mykey --keyring ~/.gnupg/secring.gpg\n```\n" +
		"If `--keyring` is not specified, Helm usually defaults to the public keyring " +
		"unless your environment is otherwise configured."

	return docs
}

func GetHelmPullDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Retrieve a package from a package repository, and download it locally.\n\n" +
		"This is useful for fetching packages to inspect, modify, or repackage. It can " +
		"also be used to perform cryptographic verification of a chart without installing " +
		"the chart.\n\n" +
		"There are options for unpacking the chart after download. This will create a " +
		"directory for the chart and uncompress into that directory.\n\n" +
		"If the `--verify` flag is specified, the requested chart **must** have a provenance " +
		"file, and **must** pass the verification process. Failure in any part of this will " +
		"result in an error, and the chart will not be saved locally."

	return docs
}
