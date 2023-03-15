package kubectl

import (
	"fmt"
	"path"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/werf/werf/cmd/werf/docs/structs"
)

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

func GetWhoamiDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Experimental: Check self subject attributes."

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

func GetCompletionDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Output shell completion code for the specified shell (Bash, Zsh, Fish, or PowerShell).\n" +
		"The shell code must be evaluated to provide interactive " +
		"completion of `kubectl` commands. This can be done by sourcing it from " +
		"the `.bash_profile`.\n\n" +
		"Detailed instructions on how to do this are available here:\n" +
		"* for macOS: " +
		"https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/#enable-shell-autocompletion;\n" +
		"* for linux: " +
		"https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion;\n\n" +
		"* for windows: " +
		"https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/#enable-shell-autocompletion;\n\n" +
		"> **Note for Zsh users**: Zsh completions are only supported in versions of Zsh >= 5.2."

	return docs
}

func GetConfigDocs(pathOptions *clientcmd.PathOptions) structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Modify kubeconfig files using subcommands like `kubectl config set current-context my-context`\n\n" +
		"The loading order follows these rules:\n" +
		"1. If the --" + pathOptions.ExplicitFileFlag + " flag is set, then only that file is loaded. " +
		"The flag may only be set once and no merging takes place.\n" +
		"2. If $" + pathOptions.EnvVar + " environment variable is set, then it is used as a list " +
		"of paths (normal path delimiting rules for your system). These paths are merged. When a value is " +
		"modified, it is modified in the file that defines the stanza. When a value is created, it is created " +
		"in the first file that exists. If no files in the chain exist, then it creates the last file in the list.\n" +
		"3. Otherwise, " + path.Join("${HOME}", pathOptions.GlobalFileSubpath) + " " +
		"is used and no merging takes place."

	return docs
}

func GetConfigCurrentContextDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display the current-context."

	return docs
}

func GetConfigDeleteClusterDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Delete the specified cluster from the kubeconfig."

	return docs
}

func GetConfigDeleteContextDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Delete the specified context from the kubeconfig."

	return docs
}

func GetConfigDeleteUserDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Delete the specified user from the kubeconfig."

	return docs
}

func GetConfigGetClustersDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display clusters defined in the kubeconfig."

	return docs
}

func GetConfigGetContextsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display one or many contexts from the kubeconfig file."

	return docs
}

func GetConfigGetUsersDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display users defined in the kubeconfig."

	return docs
}

func GetConfigRenameContextDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Renames a context from the kubeconfig file.\n" +
		"* `CONTEXT_NAME` is the context name that you want to change.\n" +
		"* `NEW_NAME` is the new name you want to set.\n\n" +
		"> **Note**: If the context being renamed is the `current-context`, this field will also be updated."

	return docs
}

func GetConfigSetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set an individual value in a kubeconfig file.\n" +
		"* `PROPERTY_NAME` is a dot delimited name where each token represents either an attribute " +
		"name or a map key.  Map keys may not contain dots.\n" +
		"* `PROPERTY_VALUE` is the new value you want to set. Binary fields such " +
		"as `certificate-authority-data` expect a base64 encoded string unless the `--set-raw-bytes` flag is used.\n\n" +
		"Specifying an attribute name that already exists will merge new fields on top of existing values."

	return docs
}

func GetConfigSetClusterDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set a cluster entry in kubeconfig.\n\n" +
		"Specifying a name that already exists will merge new fields on top of existing values for those fields."

	return docs
}

func GetConfigSetContextDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set a context entry in kubeconfig.\n\n" +
		"Specifying a name that already exists will merge new fields on top of existing values for those fields."

	return docs
}

func GetConfigSetCredentialsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = fmt.Sprintf("Set a user entry in kubeconfig.\n\n"+
		"Specifying a name that already exists will merge new fields on top of existing values:\n"+
		"* Client-certificate flags: `--%v=certfile`, `--%v=keyfile`;\n"+
		"* Bearer token flags: `--%v=bearer_token`;\n"+
		"* Basic auth flags: `--%v=basic_user`, `--%v=basic_password`.\n\n"+
		"Bearer token and basic auth are mutually exclusive.",
		clientcmd.FlagCertFile, clientcmd.FlagKeyFile, clientcmd.FlagBearerToken,
		clientcmd.FlagUsername, clientcmd.FlagPassword)

	return docs
}

func GetConfigUnsetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Unset an individual value in a kubeconfig file.\n\n" +
		"`PROPERTY_NAME` is a dot delimited name where each token represents either an attribute name " +
		"or a map key. Map keys may not contain dots."

	return docs
}

func GetConfigUseContextDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set the current-context in a kubeconfig file."

	return docs
}

func GetConfigViewDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display merged kubeconfig settings or a specified kubeconfig file.\n\n" +
		"You can use `--output jsonpath={...}` to extract specific values using a jsonpath expression."

	return docs
}

func GetCordonDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Mark node as unschedulable."

	return docs
}

func GetCpDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Copy files and directories to and from containers."

	return docs
}

func GetCreateDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a resource from a file or from stdin.\n\n" +
		"JSON and YAML formats are accepted."

	return docs
}

func GetCreateTokenDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Request a service account token."

	return docs
}

func GetCreateClusterRoleDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a cluster role."

	return docs
}

func GetCreateClusterRoleBindingDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a cluster role binding for a particular cluster role."

	return docs
}

func GetCreateConfigMapDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a config map based on a file, directory, or specified literal value.\n\n" +
		"A single config map may package one or more key/value pairs.\n\n" +
		"When creating a config map based on a file, the key will default to the basename of the file, " +
		"and the value will default to the file content.  If the basename is an invalid key, you may specify " +
		"an alternate key.\n\n" +
		"When creating a config map based on a directory, each file whose basename is a valid key in the " +
		"directory will be packaged into the config map.  Any directory entries except regular files are " +
		"ignored (e.g. subdirectories, symlinks, devices, pipes, etc)."

	return docs
}

func GetCreateCronJobDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a cron job with the specified name."

	return docs
}

func GetCreateDeploymentDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a deployment with the specified name."

	return docs
}

func GetCreateIngressDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create an ingress with the specified name."

	return docs
}

func GetCreateJobDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a job with the specified name."

	return docs
}

func GetCreateNamespaceDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a namespace with the specified name."

	return docs
}

func GetCreatePodDisruptionBudgetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a pod disruption budget with the specified name, selector, " +
		"and desired minimum available pods."

	return docs
}

func GetCreatePriorityClassDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a priority class with the specified name, value, `globalDefault` and description."

	return docs
}

func GetCreateQuotaDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a resource quota with the specified name, hard limits, and optional scopes."

	return docs
}

func GetCreateRoleDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a role with single rule."

	return docs
}

func GetCreateRoleBindingDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a role binding for a particular role or cluster role."

	return docs
}

func GetCreateSecretDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a secret using specified subcommand."

	return docs
}

func GetCreateSecretDockerRegistryDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a new secret for use with Docker registries.\n\n" +
		"Dockercfg secrets are used to authenticate against Docker registries.\n\n" +
		"When using the Docker command line to push images, you can authenticate to a given " +
		"registry by running:\n" +
		"```\n$ docker login DOCKER_REGISTRY_SERVER --username=DOCKER_USER --password=DOCKER_PASSWORD " +
		"--email=DOCKER_EMAIL\n```\n" +
		"That produces a `~/.dockercfg` file that is used by subsequent `docker push` and `docker pull` " +
		"commands to authenticate to the registry. The email address is optional.\n\n" +
		"When creating applications, you may have a Docker registry that requires authentication. " +
		"In order for the nodes to pull images on your behalf, they must have the credentials. " +
		"You can provide this information by creating a dockercfg secret and attaching it to your service account."

	return docs
}

func GetCreateSecretGenericDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a secret based on a file, directory, or specified literal value.\n\n" +
		"A single secret may package one or more key/value pairs.\n\n" +
		"When creating a secret based on a file, the key will default to the basename of the file, " +
		"and the value will default to the file content. If the basename is an invalid key or you wish " +
		"to chose your own, you may specify an alternate key.\n\n" +
		"When creating a secret based on a directory, each file whose basename is a valid key in " +
		"the directory will be packaged into the secret. Any directory entries except regular " +
		"files are ignored (e.g. subdirectories, symlinks, devices, pipes, etc)."

	return docs
}

func GetCreateSecretTLSDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a TLS secret from the given public/private key pair.\n\n" +
		"The public/private key pair must exist beforehand. The public key certificate " +
		"must be `.PEM` encoded and match the given private key."

	return docs
}

func GetCreateServiceDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a service using a specified subcommand."

	return docs
}

func GetCreateServiceClusterIPDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a ClusterIP service with the specified name."

	return docs
}

func GetCreateServiceExternalNameDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create an ExternalName service with the specified name.\n\n" +
		"ExternalName service references to an external DNS address instead of " +
		"only pods, which will allow application authors to reference services " +
		"that exist off platform, on other clusters, or locally."
	return docs
}

func GetCreateServiceLoadBalancerDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a LoadBalancer service with the specified name."

	return docs
}

func GetCreateServiceNodePortDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a NodePort service with the specified name."

	return docs
}

func GetCreateServiceAccountDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create a service account with the specified name."

	return docs
}

func GetDebugDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Debug cluster resources using interactive debugging containers.\n\n" +
		"`debug` provides automation for common debugging tasks for cluster objects identified by " +
		"resource and name. Pods will be used by default if no resource is specified.\n\n" +
		"The action taken by `debug` varies depending on what resource is specified. Supported " +
		"actions include:\n" +
		"* Workload: Create a copy of an existing pod with certain attributes changed, " +
		"for example changing the image tag to a new version.\n" +
		"* Workload: Add an ephemeral container to an already running pod, for example to add " +
		"debugging utilities without restarting the pod.\n" +
		"* Node: Create a new pod that runs in the node's host namespaces and can access " +
		"the node's filesystem."

	return docs
}

func GetDeleteDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Delete resources by file names, stdin, resources and names, " +
		"or by resources and label selector.\n\n" +
		"JSON and YAML formats are accepted. Only one type of argument may be specified: " +
		"file names, resources and names, or resources and label selector. Some resources, " +
		"such as pods, support graceful deletion. These resources define a default period " +
		"before they are forcibly terminated (the grace period) but you may override that value " +
		"with the `--grace-period` flag, or pass `--now` to set a grace-period of `1`. Because these " +
		"resources often represent entities in the cluster, deletion may not be acknowledged " +
		"immediately. If the node hosting a pod is down or cannot reach the API server, termination " +
		"may take significantly longer than the grace period. To force delete a resource, you must " +
		"specify the `--force` flag. **Note**: only a subset of resources support graceful deletion. " +
		"In absence of the support, the `--grace-period` flag is ignored.\n\n" +
		"**IMPORTANT**: Force deleting pods does not wait for confirmation that the pod's processes " +
		"have been terminated, which can leave those processes running until the node detects " +
		"the deletion and completes graceful deletion. If your processes use shared storage or " +
		"talk to a remote API and depend on the name of the pod to identify themselves, force " +
		"deleting those pods may result in multiple processes running on different machines using " +
		"the same identification which may lead to data corruption or inconsistency. Only force " +
		"delete pods when you are sure the pod is terminated, or if your application can tolerate " +
		"multiple copies of the same pod running at once. Also, if you force delete pods, the " +
		"scheduler may place new pods on those nodes before the node has released those resources " +
		"and causing those pods to be evicted immediately.\n\n" +
		"Note that the delete command does NOT do resource version checks, so if someone " +
		"submits an update to a resource right when you submit a delete, their update will " +
		"be lost along with the rest of the resource."

	return docs
}

func GetDescribeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Show details of a specific resource or group of resources.\n\n" +
		"Print a detailed description of the selected resources, including related resources " +
		"such as events or controllers. You may select a single object by name, all objects of " +
		"that type, provide a name prefix, or label selector. For example: " +
		"`$ kubectl describe TYPE NAME_PREFIX` will first check for an exact match on " +
		"`TYPE` and `NAME_PREFIX`. If no such resource exists, it will output details for every " +
		"resource that has a name prefixed with `NAME_PREFIX`.\n\n" +
		"Use `kubectl api-resources` for a complete list of supported resources."

	return docs
}

func GetDiffDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Diff configurations specified by file name or stdin between the current online " +
		"configuration, and the configuration as it would be if applied.\n\n" +
		"The output is always YAML.\n\n" +
		"`KUBECTL_EXTERNAL_DIFF` environment variable can be used to select your own " +
		"`diff` command. Users can use external commands with params too, example: " +
		"`KUBECTL_EXTERNAL_DIFF=colordiff -N -u`.\n\n" +
		"By default, the `diff` command available in your path will be run with the `-u` " +
		"(unified diff) and `-N` (treat absent files as empty) options.\n\n" +
		"Exit status:\n" +
		"* `0` – No differences were found.\n" +
		"* `1` – Differences were found.\n" +
		"* `>1` – Kubectl or diff failed with an error.\n" +
		"**Note**: `KUBECTL_EXTERNAL_DIFF`, if used, is expected to follow that convention."

	return docs
}

func GetDrainDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Drain node in preparation for maintenance.\n\n" +
		"The given node will be marked unschedulable to prevent new pods from arriving. " +
		"`drain` evicts the pods if the API server supports " +
		"eviction (https://kubernetes.io/docs/concepts/workloads/pods/disruptions/). " +
		"Otherwise, it will use normal `DELETE` to delete the pods. The `drain` evicts or " +
		"deletes all pods except mirror pods (which cannot be deleted through the API server). " +
		"deletes all pods except mirror pods (which cannot be deleted through the API server). " +
		"If there are daemon set-managed pods, drain will not proceed without " +
		"`--ignore-daemonsets`, and regardless it will not delete any daemon set-managed " +
		"pods, because those pods would be immediately replaced by the daemon set controller, " +
		"which ignores unschedulable markings. If there are any pods that are neither mirror " +
		"pods nor managed by a replication controller, replica set, daemon set, stateful set, " +
		"or job, then drain will not delete any pods unless you use `--force`. `--force` will " +
		"also allow deletion to proceed if the managing resource of one or more pods is missing.\n\n" +
		"`drain` waits for graceful termination. You should not operate on the machine until " +
		"the command completes.\n\n" +
		"When you are ready to put the node back into service, use `kubectl uncordon`, which " +
		"will make the node schedulable again.\n\n" +
		"You can view the workflow here: https://kubernetes.io/images/docs/kubectl_drain.svg"

	return docs
}

func GetEditDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Edit a resource from the default editor.\n\n" +
		"The edit command allows you to directly edit any API resource you can retrieve via the " +
		"command-line tools. It will open the editor defined by your `KUBE_EDITOR`, or `EDITOR` " +
		"environment variables, or fall back to `vi` for Linux or `notepad` for Windows. " +
		"You can edit multiple objects, although changes are applied one at a time. The command " +
		"accepts file names as well as command-line arguments, although the files you point to must " +
		"be previously saved versions of resources.\n\n" +
		"Editing is done with the API version used to fetch the resource. To edit using a specific " +
		"API version, fully-qualify the resource, version, and group.\n\n" +
		"The default format is YAML. To edit in JSON, specify `-o json`.\n\n" +
		"The flag `--windows-line-endings` can be used to force Windows line endings, otherwise the " +
		"default for your operating system will be used.\n\n" +
		"In the event an error occurs while updating, a temporary file will be created on disk " +
		"that contains your unapplied changes. The most common error when updating a resource " +
		"is another editor changing the resource on the server. When this occurs, you will have " +
		"to apply your changes to the newer version of the resource, or update your temporary " +
		"saved copy to include the latest resource version."

	return docs
}

func GetExecDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Execute a command in a container."

	return docs
}

func GetExplainDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "List the fields for supported resources.\n\n" +
		"This command describes the fields associated with each supported API resource. " +
		"Fields are identified via a simple JSON Path identifier:\n" +
		"```\n<type>.<fieldName>[.<fieldName>]\n```\n" +
		"Add the `--recursive` flag to display all of the fields at once without descriptions. " +
		"Information about each field is retrieved from the server in OpenAPI format.\n\n" +
		"Use `kubectl api-resources` for a complete list of supported resources."

	return docs
}

func GetExposeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Expose a resource as a new Kubernetes service.\n\n" +
		"Looks up a deployment, service, replica set, replication controller or pod by name " +
		"and uses the selector for that resource as the selector for a new service on the " +
		"specified port. A deployment or replica set will be exposed as a service only if " +
		"its selector is convertible to a selector that service supports, i.e. when the " +
		"selector contains only the matchLabels component. Note that if no port is specified " +
		"via `--port` and the exposed resource has multiple ports, all will be re-used by the " +
		"new service. Also if no labels are specified, the new service will re-use the " +
		"labels from the resource it exposes.\n\n" +
		"Possible resources include (case insensitive):\n" +
		"* pod (po),\n" +
		"* service (svc),\n" +
		"* replicationcontroller (rc),\n" +
		"* deployment (deploy),\n" +
		"* replicaset (rs)."

	return docs
}

func GetGetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display one or many resources.\n\n" +
		"Prints a table of the most important information about the specified resources. " +
		"You can filter the list using a label selector and the `--selector` flag. If the " +
		"desired resource type is namespaced you will only see results in your current " +
		"namespace unless you pass `--all-namespaces`.\n\n" +
		"By specifying the output as `template` and providing a Go template as the value " +
		"of the `--template` flag, you can filter the attributes of the fetched resources.\n\n" +
		"Use `kubectl api-resources` for a complete list of supported resources."

	return docs
}

func GetKustomizeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Build a set of KRM resources using a `kustomization.yaml` file. The " +
		"`DIR` argument must be a path to a directory containing `kustomization.yaml`, or a " +
		"git repository URL with a path suffix specifying same with respect to the " +
		"repository root. If `DIR` is omitted, `.` is assumed."

	return docs
}

func GetLabelDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the labels on a resource:\n" +
		"* A label key and value must begin with a letter or number, and may contain letters, " +
		"numbers, hyphens, dots, and underscores, up to `%[1]d` characters each.\n" +
		"* Optionally, the key can begin with a DNS subdomain prefix and a single `/`, like " +
		"`example.com/my-app`.\n" +
		"* If `--overwrite` is true, then existing labels can be overwritten, otherwise " +
		"attempting to overwrite a label will result in an error.\n" +
		"* If `--resource-version` is specified, then updates will use this resource " +
		"version, otherwise the existing resource-version will be used."

	return docs
}

func GetLogsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Print the logs for a container in a pod or specified resource.\n\n" +
		"If the pod has only one container, the container name is optional."

	return docs
}

func GetOptionsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Print the list of flags inherited by all commands."

	return docs
}

func GetPatchDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update fields of a resource using strategic merge patch, " +
		"a JSON merge patch, or a JSON patch.\n\n" +
		"JSON and YAML formats are accepted."

	return docs
}

func GetPluginDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Provides utilities for interacting with plugins.\n\n" +
		"Plugins provide extended functionality that is not part of the major " +
		"command-line distribution. Please refer to the documentation and examples for " +
		"more information about how write your own plugins.\n\n" +
		"The easiest way to discover and install plugins is via the kubernetes " +
		"sub-project krew. To install krew, visit " +
		"https://krew.sigs.k8s.io/docs/user-guide/setup/install/."

	return docs
}

func GetPluginListDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "List all available plugin files on a user's `PATH`.\n\n" +
		"Available plugin files are those that are:\n" +
		"* executable;\n" +
		"* anywhere on the user's `PATH`;\n" +
		"* begin with `kubectl-`."

	return docs
}

func GetPortForwardDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Forward one or more local ports to a pod.\n\n" +
		"Use resource `type`/`name` such as `deployment`/`mydeployment` to select a pod. " +
		"Resource `type` defaults to `pod` if omitted.\n\n" +
		"If there are multiple pods matching the criteria, a pod will be selected automatically. " +
		"The forwarding session ends when the selected pod terminates, and a rerun of the command is needed " +
		"to resume forwarding."

	return docs
}

func GetProxyDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Creates a proxy server or application-level gateway between localhost and " +
		"the Kubernetes API server. It also allows serving static content over specified " +
		"HTTP path. All incoming data enters through one port and gets forwarded to " +
		"the remote Kubernetes API server port, except for the path matching the static content path."

	return docs
}

func GetReplaceDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Replace a resource by file name or stdin.\n\n" +
		"JSON and YAML formats are accepted. If replacing an existing resource, the " +
		"complete resource spec must be provided. This can be obtained by `$ kubectl get TYPE NAME -o yaml`."

	return docs
}

func GetRolloutDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Manage the rollout of a resource.\n\n" +
		"Valid resource types include:\n" +
		"* deployments;\n" +
		"* daemonsets;\n" +
		"* statefulsets."

	return docs
}

func GetRolloutHistoryDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "View previous rollout revisions and configurations."

	return docs
}

func GetRolloutPauseDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Mark the provided resource as paused.\n\n" +
		"Paused resources will not be reconciled by a controller. " +
		"Use `kubectl rollout resume` to resume a paused resource. " +
		"Currently only deployments support being paused."

	return docs
}

func GetRolloutResumeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Resume a paused resource.\n\n" +
		"Paused resources will not be reconciled by a controller. By resuming a " +
		"resource, we allow it to be reconciled again. " +
		"Currently only deployments support being resumed."

	return docs
}

func GetRolloutUndoDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Roll back to a previous rollout."

	return docs
}

func GetRolloutStatusDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Show the status of the rollout.\n\n" +
		"By default `rollout status` will watch the status of the latest rollout " +
		"until it's done. If you don't want to wait for the rollout to finish then " +
		"you can use `--watch=false`. Note that if a new rollout starts in-between, then " +
		"`rollout status` will continue watching the latest revision. If you want to " +
		"pin to a specific revision and abort if it is rolled over by another revision, " +
		"use `--revision=N` where `N` is the revision you need to watch for."

	return docs
}

func GetRolloutRestartDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Restart a resource.\n\n" +
		"Resource rollout will be restarted."

	return docs
}

func GetRunDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Create and run a particular image in a pod."

	return docs
}

func GetScaleDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set a new size for a deployment, replica set, replication controller, " +
		"or stateful set.\n\n" +
		"Scale also allows users to specify one or more preconditions for the scale action.\n\n" +
		"If `--current-replicas` or `--resource-version` is specified, it is validated before the " +
		"scale is attempted, and it is guaranteed that the precondition holds true when the " +
		"scale is sent to the server."

	return docs
}

func GetSetDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Configure application resources.\n\n" +
		"These commands help you make changes to existing application resources."

	return docs
}

func GetSetImageDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update existing container image(s) of resources.\n\n" +
		"Possible resources include (case insensitive):\n" +
		"* pod (po),\n" +
		"* replicationcontroller (rc),\n" +
		"* deployment (deploy),\n" +
		"* daemonset (ds),\n" +
		"* statefulset (sts),\n" +
		"* cronjob (cj),\n" +
		"* replicaset (rs)."

	return docs
}

func GetSetResourceDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Specify compute resource requirements (CPU, memory) for any resource that " +
		"defines a pod template.  If a pod is successfully scheduled, it is guaranteed the " +
		"amount of resource requested, but may burst up to its specified limits.\n\n" +
		"For each compute resource, if a limit is specified and a request is omitted, the " +
		"request will default to the limit.\n\n" +
		"Possible resources include (case insensitive): kubectl."

	return docs
}

func GetSetSelectorDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Set the selector on a resource. Note that the new selector will overwrite " +
		"the old selector if the resource had one prior to the invocation of `set selector`.\n\n" +
		"A selector must begin with a letter or number, and may contain letters, numbers, hyphens, " +
		"dots, and underscores, up to 63 characters. If `--resource-version` is specified, then " +
		"updates will use this resource version, otherwise the existing resource-version will be used.\n\n" +
		"**Note**: currently selectors can only be set on Service objects."

	return docs
}

func GetSetSubjectDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the user, group, or service account in a role binding or cluster role binding."

	return docs
}

func GetSetServiceAccountDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the service account of pod template resources.\n\n" +
		"Possible resources (case insensitive) can be:\n" +
		"* replicationcontroller (rc),\n" +
		"* deployment (deploy),\n" +
		"* daemonset (ds),\n" +
		"* job,\n" +
		"* replicaset (rs),\n" +
		"* statefulset."

	return docs
}

func GetSetEnvDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update environment variables on a pod template.\n\n" +
		"List environment variable definitions in one or more pods, pod templates. " +
		"Add, update, or remove container environment variable definitions in one or " +
		"more pod templates (within replication controllers or deployment configurations). " +
		"View or modify the environment variable definitions on all containers in the " +
		"specified pods or pod templates, or just those that match a wildcard.\n\n" +
		"If `--env -` is passed, environment variables can be read from STDIN using the standard env " +
		"syntax.\n\n" +
		"Possible resources include (case insensitive):\n" +
		"* pod (po),\n" +
		"* replicationcontroller (rc),\n" +
		"* deployment (deploy),\n" +
		"* daemonset (ds),\n" +
		"* statefulset (sts),\n" +
		"* cronjob (cj),\n" +
		" replicaset (rs)."

	return docs
}

func GetTaintDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Update the taints on one or more nodes.\n\n" +
		"* A taint consists of a key, value, and effect. As an argument here, " +
		"it is expressed as `key=value:effect`.\n" +
		"* The key must begin with a letter or number, and may contain letters, " +
		"numbers, hyphens, dots, and underscores, up to 253 characters.\n" +
		"* Optionally, the key can begin with a DNS subdomain prefix and a single `/`, " +
		"like `example.com/my-app`.\n" +
		"* The value is optional. If given, it must begin with a letter or number, " +
		"and may contain letters, numbers, hyphens, dots, and underscores, up " +
		"to 63 characters.\n" +
		"* The effect must be `NoSchedule`, `PreferNoSchedule` or `NoExecute`.\n" +
		"* Currently taint can only apply to node."

	return docs
}

func GetTopDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display Resource (CPU/Memory) usage.\n\n" +
		"The top command allows you to see the resource consumption for nodes or pods.\n\n" +
		"This command requires Metrics Server to be correctly configured and working on the server."

	return docs
}

func GetTopNodeDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display resource (CPU/memory) usage of nodes.\n\n" +
		"The `top-node` command allows you to see the resource consumption of nodes."

	return docs
}

func GetTopPodDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display resource (CPU/memory) usage of pods.\n\n" +
		"The `top pod` command allows you to see the resource consumption of pods.\n\n" +
		"Due to the metrics pipeline delay, they may be unavailable for a few minutes since pod creation."

	return docs
}

func GetUncordonDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Mark node as schedulable."

	return docs
}

func GetVersionDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Print the client and server version information for the current context."

	return docs
}

func GetWaitDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "**Experimental**: Wait for a specific condition on one or many resources.\n\n" +
		"The command takes multiple resources and waits until the specified condition is seen " +
		"in the `Status` field of every given resource.\n\n" +
		"Alternatively, the command can wait for the given set of resources to be deleted " +
		"by providing the `delete` keyword as the value to the `--for` flag.\n\n" +
		"A successful message will be printed to stdout indicating when the specified " +
		"condition has been met. You can use `-o` option to change to output destination."

	return docs
}

func GetEventsDocs() structs.DocsStruct {
	var docs structs.DocsStruct

	docs.LongMD = "Display events.\n\n" +
		"Prints a table of the most important information about events.\n\n" +
		"You can request events for a namespace, for all namespace, or filtered to only those " +
		"pertaining to a specified resource."

	return docs
}
