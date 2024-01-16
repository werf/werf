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

func GetHelmRollbackDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command rolls back a release to a previous revision.\n\n" +
		"The first argument of the rollback command is the name of a release, and the " +
		"second is a revision (version) number. If this argument is omitted, it will " +
		"roll back to the previous release.\n\n" +
		"To see revision numbers, run `helm history RELEASE`."

	return docs
}

func GetHelmStatusDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command shows the status of a named release. " +
		"The status consists of:\n" +
		"- last deployment time;\n" +
		"- K8s namespace in which the release lives;\n" +
		"- state of the release (can be: `unknown`, `deployed`, `uninstalled`, " +
		"`superseded`, `failed`, `uninstalling`, `pending-install`, `pending-upgrade` or `pending-rollback`);\n" +
		"- revision of the release;\n" +
		"- description of the release (can be completion message or error message, need to enable `--show-desc`);\n" +
		"- list of resources that this release consists of, sorted by kind;\n" +
		"- details on last test suite run, if applicable;\n" +
		"- additional notes provided by the chart.\n"

	return docs
}

func GetHelmUninstallDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command takes a release name and uninstalls the release.\n\n" +
		"It removes all of the resources associated with the last release of the chart " +
		"as well as the release history, freeing it up for future use.\n\n" +
		"Use the `--dry-run` flag to see which releases will be uninstalled without actually " +
		"uninstalling them."

	return docs
}

func GetHelmUpgradeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command upgrades a release to a new version of a chart.\n\n" +
		"The upgrade arguments must be a release and chart. The chart " +
		"argument can be either: a chart reference(`example/mariadb`), a path to a chart directory, " +
		"a packaged chart, or a fully qualified URL. For chart references, the latest " +
		"version will be specified unless the `--version` flag is set.\n\n" +
		"To override values in a chart, use either the `--values` flag and pass in a file " +
		"or use the `--set` flag and pass configuration from the command line, to force string " +
		"values, use `--set-string`. You can use `--set-file` to set individual " +
		"values from a file when the value itself is too long for the command line " +
		"or is dynamically generated.\n\n" +
		"You can specify the `--values`/`-f` flag multiple times. The priority will be given to the " +
		"last (right-most) file specified. For example, if both `myvalues.yaml` and `override.yaml` " +
		"contained a key called `Test`, the value set in `override.yaml` would take precedence:\n" +
		"```\n$ helm upgrade -f myvalues.yaml -f override.yaml redis ./redis\n```\n" +
		"You can specify the `--set` flag multiple times. The priority will be given to the " +
		"last (right-most) set specified. For example, if both `bar` and `newbar` values are " +
		"set for a key called `foo`, the `newbar` value would take precedence:\n" +
		"```\n$ helm upgrade --set foo=bar --set foo=newbar redis ./redis\n```\n"

	return docs
}

func GetHelmVerifyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Verify that the given chart has a valid provenance file.\n\n" +
		"Provenance files provide cryptographic verification that a chart has not been " +
		"tampered with, and was packaged by a trusted provider.\n\n" +
		"This command can be used to verify a local chart. Several other commands provide " +
		"`--verify` flags that run the same validation. To generate a signed package, use " +
		"the `helm package --sign` command."

	return docs
}

func GetHelmVersionDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Show the version for Helm.\n\n" +
		"This will print a representation the version of Helm. " +
		"The output will look something like this:\n" +
		"```\nversion.BuildInfo{Version:\"v3.2.1\", GitCommit:\"fe51cd1e31e6a202cba7dead9552a6d418ded79a\", " +
		"GitTreeState:\"clean\", GoVersion:\"go1.13.10\"}\n```\n" +
		"- `Version` is the semantic version of the release;\n" +
		"- `GitCommit` is the SHA for the commit that this version was built from;\n" +
		"- `GitTreeState` is `clean` if there are no local code changes when this binary was" +
		"  built, and `dirty` if the binary was built from locally modified code;\n" +
		"- `GoVersion` is the version of Go that was used to compile Helm.\n\n" +
		"When using the `--template` flag the following properties are available to use in " +
		"the template:\n" +
		"- `.Version` contains the semantic version of Helm;\n" +
		"- `.GitCommit` is the git commit;\n" +
		"- `.GitTreeState` is the state of the git tree when Helm was built;\n" +
		"- `.GoVersion` contains the version of Go that Helm was compiled with.\n\n" +
		"For example, `--template='Version: {{.Version}}'` outputs `'Version: v3.2.1'`."

	return docs
}

func GetHelmDependencyBuildDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Build out the `charts/` directory from the `Chart.lock` file.\n\n" +
		"Build is used to reconstruct a chart's dependencies to the state specified in " +
		"the lock file. This will not re-negotiate dependencies, as `helm dependency update` " +
		"does.\n\n" +
		"If no lock file is found, `helm dependency build` will mirror the behavior " +
		"of `helm dependency update`.\n"

	return docs
}

func GetHelmDependencyUpdateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the on-disk dependencies to mirror `Chart.yaml`.\n\n" +
		"This command verifies that the required charts, as expressed in `Chart.yaml`, " +
		"are present in `charts/` and are at an acceptable version. It will pull down " +
		"the latest charts that satisfy the dependencies, and clean up old dependencies.\n\n" +
		"On successful update, this will generate a lock file that can be used to " +
		"rebuild the dependencies to an exact version.\n\n" +
		"Dependencies are not required to be represented in `Chart.yaml`. For that " +
		"reason, an update command will not remove charts unless they are (a) present " +
		"in the `Chart.yaml` file, but (b) at the wrong version.\n"

	return docs
}

func GetHelmDependencyListDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "List all of the dependencies declared in a chart.\n\n" +
		"This can take chart archives and chart directories as input. It will not alter " +
		"the contents of a chart.\n\n" +
		"This will produce an error if the chart cannot be loaded.\n"

	return docs
}

func GetHelmGetHooksDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command downloads hooks for a given release.\n\n" +
		"Hooks are formatted in YAML and separated by the YAML `---\\n` separator."

	return docs
}

func GetHelmGetAllDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command prints a human readable collection of information about the " +
		"notes, hooks, supplied values, and generated manifest file of the given release."

	return docs
}

func GetHelmGetValuesDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command downloads a values file for a given release."

	return docs
}

func GetHelmGetManifestDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command fetches the generated manifest for a given release.\n\n" +
		"A manifest is a YAML-encoded representation of the Kubernetes resources that " +
		"were generated from this release's chart(s). If a chart is dependent on other " +
		"charts, those resources will also be included in the manifest."

	return docs
}

func GetHelmGetNotesDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command shows notes provided by the chart of a named release."

	return docs
}

func GetHelmGetMetadataDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = ""

	return docs
}

func GetHelmPluginListDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "List installed Helm plugins."

	return docs
}

func GetHelmPluginUninstallDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Uninstall one or more Helm plugins."

	return docs
}

func GetHelmPluginUpdateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update one or more Helm plugins."

	return docs
}

func GetHelmPluginInstallDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command allows you to install a plugin from a URL to a VCS repo or a local path."

	return docs
}

func GetHelmRepoAddDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Add a chart repository."

	return docs
}

func GetHelmRepoIndexDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Read the current directory and generate an index file based on the charts found.\n\n" +
		"This tool is used for creating an `index.yaml` file for a chart repository. To " +
		"set an absolute URL to the charts, use `--url` flag.\n\n" +
		"To merge the generated index with an existing index file, use the `--merge` " +
		"flag. In this case, the charts found in the current directory will be merged " +
		"into the existing index, with local charts taking priority over existing charts.\n"

	return docs
}

func GetHelmRepoListDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "List chart repositories."

	return docs
}

func GetHelmRepoRemoveDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Remove one or more chart repositories."

	return docs
}

func GetHelmRepoUpdateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update gets the latest information about charts from the respective chart repositories. " +
		"Information is cached locally, where it is used by commands like `helm search`.\n\n" +
		"You can optionally specify a list of repositories you want to update:\n" +
		"```\n$ helm repo update <repo_name> ...\n```\n" +
		"To update all the repositories, use `helm repo update`."

	return docs
}

func GetHelmSearchHubDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Search for Helm charts in the Artifact Hub " +
		"or your own hub instance.\n\n" +
		"Artifact Hub is a web-based application that enables finding, installing, and " +
		"publishing packages and configurations for CNCF projects, including publicly " +
		"available distributed charts Helm charts. It is a Cloud Native Computing " +
		"Foundation sandbox project. You can browse the hub at https://artifacthub.io/.\n\n" +
		"The `[KEYWORD]` argument accepts either a keyword string, or quoted string of rich " +
		"query options. For rich query options documentation, see\nhttps://artifacthub.github.io" +
		"/hub/api/?urls.primaryName=Monocular%20compatible%20search%20API#/Monocular/get_api_chartsvc_v1_charts_search.\n\n" +
		"Previous versions of Helm used an instance of Monocular as the default " +
		"`endpoint`, so for backwards compatibility Artifact Hub is compatible with the " +
		"Monocular search API. Similarly, when setting " +
		"the `endpoint` flag, the specified " +
		"endpoint must also be implement a Monocular compatible search API endpoint. " +
		"Note that when specifying a Monocular instance as the `endpoint`, rich queries " +
		"are not supported. For API details, see https://github.com/helm/monocular.\n"

	return docs
}

func GetHelmSearchRepoDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Search reads through all of the repositories configured on the system, and " +
		"looks for matches. Search of these repositories uses the metadata stored on " +
		"the system.\n\n" +
		"It will display the latest stable versions of the charts found. If you " +
		"specify the `--devel` flag, the output will include pre-release versions. " +
		"If you want to search using a version constraint, use `--version`.\n\n" +
		"Examples:\n" +
		"```\n# Search for stable release versions matching the keyword 'nginx'\n" +
		"$ helm search repo nginx\n\n" +
		"# Search for release versions matching the keyword 'nginx', including pre-release versions\n" +
		"$ helm search repo nginx --devel\n\n" +
		"# Search for the latest stable release for nginx-ingress with a major version of 1\n" +
		"$ helm search repo nginx-ingress --version ^1.0.0\n```\n" +
		"Repositories are managed with `helm repo` commands.\n"

	return docs
}

func GetHelmSecretDecryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Decrypt data from standard input.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Decrypt data from standard input.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretEncryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Encrypt data from standard input.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Encrypt data from standard input.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretGenerateSecretKeyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Generate hex encryption key.
For further usage, the encryption key should be saved in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Generate hex encryption key.\n\n" +
		"For further usage, the encryption key should be saved in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretRotateSecretKeyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Regenerate Secret files with new Secret key.

Old key should be specified in the $WERF_OLD_SECRET_KEY.
New key should reside either in the $WERF_SECRET_KEY or .werf_secret_key file.

Command will extract data with the old key, generate new secret data and rewrite files:
* standard raw Secret files in the .helm/secret folder;
* standard secret Values YAML file .helm/secret-values.yaml;
* additional secret Values YAML files specified with EXTRA_SECRET_VALUES_FILE_PATH params`

	docs.LongMD = "Regenerate Secret files with new Secret key.\n\n" +
		"Old key should be specified in the `$WERF_OLD_SECRET_KEY`.\n\n" +
		"New key should reside either in the `$WERF_SECRET_KEY` or `.werf_secret_key file`.\n\n" +
		"Command will extract data with the old key, generate new Secret data and rewrite files:\n" +
		"* standard raw Secret files in the `.helm/secret folder`;\n" +
		"* standard Secret Values YAML file `.helm/secret-values.yaml`;\n" +
		"* additional Secret Values YAML files specified with `EXTRA_SECRET_VALUES_FILE_PATH` params."

	return docs
}

func GetHelmSecretFileDecryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Decrypt data from FILE_PATH or pipe.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Decrypt data from `FILE_PATH` or pipe.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretFileEditDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Edit or create new secret file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Edit or create new secret file.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretFileEncryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Encrypt data from FILE_PATH or pipe.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Encrypt data from `FILE_PATH` or pipe.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretValuesDecryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Decrypt data from FILE_PATH or pipe.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Decrypt data from `FILE_PATH` or pipe.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmSecretValuesEditDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Edit or create new secret values file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Edit or create new secret values file.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key` file."

	return docs
}

func GetHelmSecretValuesEncryptDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = `Encrypt data from FILE_PATH or pipe.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file.`

	docs.LongMD = "Encrypt data from `FILE_PATH` or pipe.\n\n" +
		"Encryption key should be in `$WERF_SECRET_KEY` or `.werf_secret_key file`."

	return docs
}

func GetHelmShowAllDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command inspects a chart (directory, file, or URL) and displays all its content " +
		"(`values.yaml`, `Chart.yaml`, `README`)."

	return docs
}

func GetHelmShowChartDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command inspects a chart (directory, file, or URL) and displays the contents " +
		"of the `Chart.yaml` file."

	return docs
}

func GetHelmShowCRDsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command inspects a chart (directory, file, or URL) and displays the contents " +
		"of the CustomResourceDefinition files."

	return docs
}

func GetHelmShowReadmeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command inspects a chart (directory, file, or URL) and displays the contents " +
		"of the README file."

	return docs
}

func GetHelmShowValuesDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "This command inspects a chart (directory, file, or URL) and displays the contents " +
		"of the `values.yaml` file."

	return docs
}
