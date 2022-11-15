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

func GetApplyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Apply a configuration to a resource by file name or stdin.\n" +
		"The resource name must be specified. This resource will be created if it doesn't exist yet. " +
		"To use `apply`, always create the resource initially with either `apply` or `create " +
		"--save-config`.\n\n" +
		"JSON and YAML formats are accepted.\n\n" +
		"Alpha Disclaimer: the `--prune` functionality is not yet complete. " +
		"Do not use unless you are aware of what the current state is. " +
		"See https://issues.k8s.io/34274."

	return docs
}

func GetApplyEditLastAppliedDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Edit the latest last-applied-configuration annotations of resources from the default editor.\n\n" +
		"The `edit-last-applied` command allows you to directly edit any API resource you can retrieve via the " +
		"command-line tools. It will open the editor defined by your `KUBE_EDITOR`, or `EDITOR` " +
		"environment variables, or fall back to `vi` for Linux or `notepad` for Windows. " +
		"You can edit multiple objects, although changes are applied one at a time. The command " +
		"accepts file names as well as command-line arguments, although the files you point to must " +
		"be previously saved versions of resources.\n\n" +
		"The default format is YAML. To edit in JSON, specify `-o json`.\n\n" +
		"The flag `--windows-line-endings` can be used to force Windows line endings, " +
		"otherwise the default for your operating system will be used.\n\n" +
		"In the event an error occurs while updating, a temporary file will be created on disk " +
		"that contains your unapplied changes. The most common error when updating a resource " +
		"is another editor changing the resource on the server. When this occurs, you will have " +
		"to apply your changes to the newer version of the resource, or update your temporary " +
		"saved copy to include the latest resource version."

	return docs
}

func GetApplySetLastAppliedDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set the latest `last-applied-configuration` annotations by setting it to match the contents of a file. " +
		"This results in the `last-applied-configuration` being updated as though `kubectl apply -f <file>` was run, " +
		"without updating any other parts of the object."

	return docs
}

func GetApplyViewLastAppliedDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "View the latest `last-applied-configuration` annotations by `type`/`name` or file.\n\n" +
		"The default output will be printed to stdout in YAML format. You can use the `-o` option " +
		"to change the output format."

	return docs
}

func GetAttachDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Attach to a process that is already running inside an existing container."

	return docs
}

func GetAuthDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Inspect authorization."

	return docs
}

func GetAuthCanIDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Check whether an action is allowed.\n\n" +
		"* `VERB` is a logical Kubernetes API verb like `get`, `list`, `watch`, `delete`, etc.\n" +
		"* `TYPE` is a Kubernetes resource. Shortcuts and groups will be resolved.\n" +
		"* `NONRESOURCEURL` is a partial URL that starts with `/`.\n" +
		"* `NAME` is the name of a particular Kubernetes resource.\n\n" +
		"This command pairs nicely with impersonation. See `--as global` flag."

	return docs
}

func GetAuthReconcileDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Reconciles rules for RBAC role, role binding, cluster role, and cluster role binding objects.\n\n" +
		"Missing objects are created, and the containing namespace is created for namespaced objects, if required.\n\n" +
		"Existing roles are updated to include the permissions in the input objects, " +
		"and remove extra permissions if `--remove-extra-permissions` is specified.\n\n" +
		"Existing bindings are updated to include the subjects in the input objects, " +
		"and remove extra subjects if `--remove-extra-subjects` is specified.\n\n" +
		"This is preferred to `apply` for RBAC resources so that semantically-aware " +
		"merging of rules and subjects is done."

	return docs
}

func GetAutoscaleDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Creates an autoscaler that automatically chooses and sets the number of pods that run in a " +
		"Kubernetes cluster.\n\n" +
		"Looks up a deployment, replica set, stateful set, or replication controller by name and creates an " +
		"autoscaler that uses the given resource as a reference.\n" +
		"An autoscaler can automatically increase or decrease number of Pods deployed within the system as needed."

	return docs
}

func GetCertificateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Modify certificate resources."

	return docs
}

func GetCertificateApproveDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Approve a certificate signing request.\n\n" +
		"kubectl certificate approve allows a cluster admin to approve a certificate " +
		"signing request (CSR). This action tells a certificate signing controller to " +
		"issue a certificate to the requestor with the attributes requested in the CSR.\n\n" +
		"> **SECURITY NOTICE**: Depending on the requested attributes, the issued certificate " +
		"can potentially grant a requester access to cluster resources or to authenticate " +
		"as a requested identity. Before approving a CSR, ensure you understand what the " +
		"signed certificate can do."

	return docs
}

func GetCertificateDenyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Deny a certificate signing request.\n\n" +
		"kubectl certificate deny allows a cluster admin to deny a certificate " +
		"signing request (CSR). This action tells a certificate signing controller to " +
		"not to issue a certificate to the requestor."

	return docs
}

func GetClusterInfoDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display addresses of the control plane and services with label `kubernetes.io/cluster-service=true`. " +
		"To further debug and diagnose cluster problems, use `kubectl cluster-info dump`."

	return docs
}

func GetClusterInfoDumpDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Dump cluster information out suitable for debugging and diagnosing cluster problems. " +
		"By default, dumps everything to stdout. You can optionally specify a directory with " +
		"`--output-directory`.  If you specify a directory, Kubernetes will " +
		"build a set of files in that directory. By default, only dumps things in the current namespace " +
		"and `kube-system` namespace, but you can switch to a different namespace with the " +
		"`--namespaces flag`, or specify `--all-namespaces` to dump all namespaces.\n\n" +
		"The command also dumps the logs of all of the pods in the cluster; these logs are dumped " +
		"into different directories based on namespace and Pod name."

	return docs
}
