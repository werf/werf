{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
**Experimental**: Wait for a specific condition on one or many resources.

The command takes multiple resources and waits until the specified condition is seen in the `Status` field of every given resource.

Alternatively, the command can wait for the given set of resources to be deleted by providing the `delete` keyword as the value to the `--for` flag.

A successful message will be printed to stdout indicating when the specified condition has been met. You can use `-o` option to change to output destination.

{{ header }} Syntax

```shell
werf kubectl wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available|--for=jsonpath='{}'[=value]] [options]
```

{{ header }} Examples

```shell
  # Wait for the pod "busybox1" to contain the status condition of type "Ready"
  kubectl wait --for=condition=Ready pod/busybox1
  
  # The default value of status condition is true; you can wait for other targets after an equal delimiter (compared after Unicode simple case folding, which is a more general form of case-insensitivity)
  kubectl wait --for=condition=Ready=false pod/busybox1
  
  # Wait for the pod "busybox1" to contain the status phase to be "Running"
  kubectl wait --for=jsonpath='{.status.phase}'=Running pod/busybox1
  
  # Wait for pod "busybox1" to be Ready
  kubectl wait --for='jsonpath={.status.conditions[?(@.type=="Ready")].status}=True' pod/busybox1
  
  # Wait for the service "loadbalancer" to have ingress.
  kubectl wait --for=jsonpath='{.status.loadBalancer.ingress}' service/loadbalancer
  
  # Wait for the pod "busybox1" to be deleted, with a timeout of 60s, after having issued the "delete" command
  kubectl delete pod/busybox1
  kubectl wait --for=delete pod/busybox1 --timeout=60s
```

{{ header }} Options

```shell
      --all=false
            Select all resources in the namespace of the specified resource types
  -A, --all-namespaces=false
            If present, list the requested object(s) across all namespaces. Namespace in current    
            context is ignored even if specified with --namespace.
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --field-selector=''
            Selector (field query) to filter on, supports `=`, `==`, and `!=`.(e.g.                 
            --field-selector key1=value1,key2=value2). The server only supports a limited number of 
            field queries per type.
  -f, --filename=[]
            identifying the resource.
      --for=''
            The condition to wait on:                                                               
            [delete|condition=condition-name[=condition-value]|jsonpath=`{JSONPath                  
            expression}`=[JSONPath value]]. The default condition-value is true.  Condition values  
            are compared after Unicode simple case folding, which is a more general form of         
            case-insensitivity.
      --local=false
            If true, annotation will NOT contact api-server but run locally.
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
  -R, --recursive=true
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2)
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
      --timeout=30s
            The length of time to wait before giving up.  Zero means check once and don`t wait,     
            negative means wait for a week.
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

