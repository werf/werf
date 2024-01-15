{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Display one or many resources.

Prints a table of the most important information about the specified resources. You can filter the list using a label selector and the `--selector` flag. If the desired resource type is namespaced you will only see results in your current namespace unless you pass `--all-namespaces`.

By specifying the output as `template` and providing a Go template as the value of the `--template` flag, you can filter the attributes of the fetched resources.

Use `kubectl api-resources` for a complete list of supported resources.

{{ header }} Syntax

```shell
werf kubectl get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags] [options]
```

{{ header }} Examples

```shell
  # List all pods in ps output format
  kubectl get pods
  
  # List all pods in ps output format with more information (such as node name)
  kubectl get pods -o wide
  
  # List a single replication controller with specified NAME in ps output format
  kubectl get replicationcontroller web
  
  # List deployments in JSON output format, in the "v1" version of the "apps" API group
  kubectl get deployments.v1.apps -o json
  
  # List a single pod in JSON output format
  kubectl get -o json pod web-pod-13je7
  
  # List a pod identified by type and name specified in "pod.yaml" in JSON output format
  kubectl get -f pod.yaml -o json
  
  # List resources from a directory with kustomization.yaml - e.g. dir/kustomization.yaml
  kubectl get -k dir/
  
  # Return only the phase value of the specified pod
  kubectl get -o template pod/web-pod-13je7 --template={{.status.phase}}
  
  # List resource information in custom columns
  kubectl get pod test-pod -o custom-columns=CONTAINER:.spec.containers[0].name,IMAGE:.spec.containers[0].image
  
  # List all replication controllers and services together in ps output format
  kubectl get rc,services
  
  # List one or more resources by their type and names
  kubectl get rc/web service/frontend pods/web-pod-13je7
  
  # List the 'status' subresource for a single pod
  kubectl get pod web-pod-13je7 --subresource status
```

{{ header }} Options

```shell
  -A, --all-namespaces=false
            If present, list the requested object(s) across all namespaces. Namespace in current    
            context is ignored even if specified with --namespace.
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --chunk-size=500
            Return large lists in chunks rather than all at once. Pass 0 to disable. This flag is   
            beta and may change in the future.
      --field-selector=''
            Selector (field query) to filter on, supports `=`, `==`, and `!=`.(e.g.                 
            --field-selector key1=value1,key2=value2). The server only supports a limited number of 
            field queries per type.
  -f, --filename=[]
            Filename, directory, or URL to files identifying the resource to get from a server.
      --ignore-not-found=false
            If the requested object does not exist the command will return exit code 0.
  -k, --kustomize=''
            Process the kustomization directory. This flag can`t be used together with -f or -R.
  -L, --label-columns=[]
            Accepts a comma separated list of labels that are going to be presented as columns.     
            Names are case-sensitive. You can also use multiple flag options like -L label1 -L      
            label2...
      --no-headers=false
            When using the default or custom-column output format, don`t print headers (default     
            print headers).
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file, custom-columns,                
            custom-columns-file, wide). See custom columns                                          
            [https://kubernetes.io/docs/reference/kubectl/#custom-columns], golang template         
            [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template               
            [https://kubernetes.io/docs/reference/kubectl/jsonpath/].
      --output-watch-events=false
            Output watch event objects when --watch or --watch-only is used. Existing objects are   
            output as initial ADDED events.
      --raw=''
            Raw URI to request from the server.  Uses the transport specified by the kubeconfig     
            file.
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2). Matching objects must satisfy all of the specified label      
            constraints.
      --server-print=true
            If true, have the server return the appropriate table output. Supports extension APIs   
            and CRDs.
      --show-kind=false
            If present, list the resource type for the requested object(s).
      --show-labels=false
            When printing, show all labels as the last column (default hide labels column)
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --sort-by=''
            If non-empty, sort list types using this field specification.  The field specification  
            is expressed as a JSONPath expression (e.g. `{.metadata.name}`). The field in the API   
            resource specified by this JSONPath expression must be an integer or a string.
      --subresource=''
            If specified, gets the subresource of the requested object. Must be one of [status      
            scale]. This flag is beta and may change in the future.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
  -w, --watch=false
            After listing/getting the requested object, watch for changes.
      --watch-only=false
            Watch for changes to the requested object(s), without listing/getting first.
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

