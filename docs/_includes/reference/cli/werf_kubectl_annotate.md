{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Update the annotations on one or more resources.

All Kubernetes objects support the ability to store additional data with the object as annotations. Annotations are key/value pairs that can be larger than labels and include arbitrary string values such as structured JSON. Tools and system extensions may use annotations to store their own data.

Attempting to set an annotation that already exists will fail unless `--overwrite` is set. If `--resource-version` is specified and does not match the current resource version on the server the command will fail.

Use `kubectl api-resources` for a complete list of supported resources.

{{ header }} Syntax

```shell
werf kubectl annotate [--overwrite] (-f FILENAME | TYPE NAME) KEY_1=VAL_1 ... KEY_N=VAL_N [--resource-version=version] [options]
```

{{ header }} Examples

```shell
  # Update pod 'foo' with the annotation 'description' and the value 'my frontend'
  # If the same annotation is set multiple times, only the last value will be applied
  kubectl annotate pods foo description='my frontend'
  
  # Update a pod identified by type and name in "pod.json"
  kubectl annotate -f pod.json description='my frontend'
  
  # Update pod 'foo' with the annotation 'description' and the value 'my frontend running nginx', overwriting any existing value
  kubectl annotate --overwrite pods foo description='my frontend running nginx'
  
  # Update all pods in the namespace
  kubectl annotate pods --all description='my frontend running nginx'
  
  # Update pod 'foo' only if the resource is unchanged from version 1
  kubectl annotate pods foo description='my frontend running nginx' --resource-version=1
  
  # Update pod 'foo' by removing an annotation named 'description' if it exists
  # Does not require the --overwrite flag
  kubectl annotate pods foo description-
```

{{ header }} Options

```shell
      --all=false
            Select all resources, in the namespace of the specified resource types.
  -A, --all-namespaces=false
            If true, check the specified action in all namespaces.
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --dry-run="none"
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
      --field-manager="kubectl-annotate"
            Name of the manager used to track field ownership.
      --field-selector=""
            Selector (field query) to filter on, supports `=`, `==`, and `!=`.(e.g.                 
            --field-selector key1=value1,key2=value2). The server only supports a limited number of 
            field queries per type.
  -f, --filename=[]
            Filename, directory, or URL to files identifying the resource to update the annotation
  -k, --kustomize=""
            Process the kustomization directory. This flag can`t be used together with -f or -R.
      --list=false
            If true, display the annotations for a given resource.
      --local=false
            If true, annotation will NOT contact api-server but run locally.
  -o, --output=""
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --overwrite=false
            If true, allow annotations to be overwritten, otherwise reject annotation updates that  
            overwrite existing annotations.
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
      --resource-version=""
            If non-empty, the annotation update will only succeed if this is the current            
            resource-version for the object. Only valid when specifying a single resource.
  -l, --selector=""
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2). Matching objects must satisfy all of the specified label      
            constraints.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=""
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
```

{{ header }} Options inherited from parent commands

```shell
      --as=""
            Username to impersonate for the operation. User could be a regular user or a service    
            account in a namespace.
      --as-group=[]
            Group to impersonate for the operation, this flag can be repeated to specify multiple   
            groups.
      --as-uid=""
            UID to impersonate for the operation.
      --cache-dir="~/.kube/cache"
            Default cache directory
      --certificate-authority=""
            Path to a cert file for the certificate authority
      --client-certificate=""
            Path to a client certificate file for TLS
      --client-key=""
            Path to a client key file for TLS
      --cluster=""
            The name of the kubeconfig cluster to use
      --context=""
            The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)
      --disable-compression=false
            If true, opt-out of response compression for all requests to the server
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --kube-config-base64=""
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kubeconfig=""
            Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or      
            $WERF_KUBECONFIG, or $KUBECONFIG). Ignored if kubeconfig passed as base64.
      --log-flush-frequency=5s
            Maximum number of seconds between log flushes
      --match-server-version=false
            Require server version to match client version
  -n, --namespace=""
            If present, the namespace scope for this CLI request
      --password=""
            Password for basic authentication to the API server
      --profile="none"
            Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)
      --profile-output="profile.pprof"
            Name of the file to write the profile to
      --request-timeout="0"
            The length of time to wait before giving up on a single server request. Non-zero values 
            should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don`t 
            timeout requests.
  -s, --server=""
            The address and port of the Kubernetes API server
      --tls-server-name=""
            Server name to use for server certificate validation. If it is not provided, the        
            hostname used to contact the server is used
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --token=""
            Bearer token for authentication to the API server
      --user=""
            The name of the kubeconfig user to use
      --username=""
            Username for basic authentication to the API server
  -v, --v=0
            number for the log level verbosity
      --vmodule=
            comma-separated list of pattern=N settings for file-filtered logging (only works for    
            the default text log format)
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

