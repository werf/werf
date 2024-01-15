{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Replace a resource by file name or stdin.

JSON and YAML formats are accepted. If replacing an existing resource, the complete resource spec must be provided. This can be obtained by `$ kubectl get TYPE NAME -o yaml`.

{{ header }} Syntax

```shell
werf kubectl replace -f FILENAME [options]
```

{{ header }} Examples

```shell
  # Replace a pod using the data in pod.json
  kubectl replace -f ./pod.json
  
  # Replace a pod based on the JSON passed into stdin
  cat pod.json | kubectl replace -f -
  
  # Update a single-container pod's image version (tag) to v4
  kubectl get pod mypod -o yaml | sed 's/\(image: myimage\):.*$/\1:v4/' | kubectl replace -f -
  
  # Force replace, delete and then re-create the resource
  kubectl replace --force -f ./pod.json
```

{{ header }} Options

```shell
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --cascade='background'
            Must be "background", "orphan", or "foreground". Selects the deletion cascading         
            strategy for the dependents (e.g. Pods created by a ReplicationController). Defaults to 
            background.
      --dry-run='none'
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
      --field-manager='kubectl-replace'
            Name of the manager used to track field ownership.
  -f, --filename=[]
            The files that contain the configurations to replace.
      --force=false
            If true, immediately remove resources from API and bypass graceful deletion. Note that  
            immediate deletion of some resources may result in inconsistency or data loss and       
            requires confirmation.
      --grace-period=-1
            Period of time in seconds given to the resource to terminate gracefully. Ignored if     
            negative. Set to 1 for immediate shutdown. Can only be set to 0 when --force is true    
            (force deletion).
  -k, --kustomize=''
            Process a kustomization directory. This flag can`t be used together with -f or -R.
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --raw=''
            Raw URI to PUT to the server.  Uses the transport specified by the kubeconfig file.
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
      --save-config=false
            If true, the configuration of current object will be saved in its annotation.           
            Otherwise, the annotation will be unchanged. This flag is useful when you want to       
            perform kubectl apply on this object in the future.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --subresource=''
            If specified, replace will operate on the subresource of the requested object. Must be  
            one of [status scale]. This flag is beta and may change in the future.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
      --timeout=0s
            The length of time to wait before giving up on a delete, zero means determine a timeout 
            from the size of the object
      --validate='strict'
            Must be one of: strict (or true), warn, ignore (or false).
            		"true" or "strict" will use a schema to validate the input and fail the request if    
            invalid. It will perform server side validation if ServerSideFieldValidation is enabled 
            on the api-server, but will fall back to less reliable client-side validation if not.
            		"warn" will warn about unknown or duplicate fields without blocking the request if    
            server-side field validation is enabled on the API server, and behave as "ignore"       
            otherwise.
            		"false" or "ignore" will not perform any schema validation, silently dropping any     
            unknown or duplicate fields.
      --wait=false
            If true, wait for resources to be gone before returning. This waits for finalizers.
```

{{ header }} Options inherited from parent commands

```shell
      --as=''
            Username to impersonate for the operation. User could be a regular user or a service    
            account in a namespace.
      --as-group=[]
            Group to impersonate for the operation, this flag can be repeated to specify multiple   
            groups.
      --as-uid=''
            UID to impersonate for the operation.
      --cache-dir='~/.kube/cache'
            Default cache directory
      --certificate-authority=''
            Path to a cert file for the certificate authority
      --client-certificate=''
            Path to a client certificate file for TLS
      --client-key=''
            Path to a client key file for TLS
      --cluster=''
            The name of the kubeconfig cluster to use
      --context=''
            The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)
      --disable-compression=false
            If true, opt-out of response compression for all requests to the server
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kubeconfig=''
            Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or      
            $WERF_KUBECONFIG, or $KUBECONFIG). Ignored if kubeconfig passed as base64.
      --match-server-version=false
            Require server version to match client version
  -n, --namespace=''
            If present, the namespace scope for this CLI request
      --password=''
            Password for basic authentication to the API server
      --profile='none'
            Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)
      --profile-output='profile.pprof'
            Name of the file to write the profile to
      --request-timeout='0'
            The length of time to wait before giving up on a single server request. Non-zero values 
            should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don`t 
            timeout requests.
  -s, --server=''
            The address and port of the Kubernetes API server
      --tls-server-name=''
            Server name to use for server certificate validation. If it is not provided, the        
            hostname used to contact the server is used
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --token=''
            Bearer token for authentication to the API server
      --user=''
            The name of the kubeconfig user to use
      --username=''
            Username for basic authentication to the API server
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

