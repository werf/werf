{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Apply a configuration to a resource by file name or stdin. The resource name must be specified. This resource will be created if it doesn&#39;t exist yet. To use &#39;apply&#39;, always create the resource initially with either &#39;apply&#39; or &#39;create --save-config&#39;.

 JSON and YAML formats are accepted.

 Alpha Disclaimer: the --prune functionality is not yet complete. Do not use unless you are aware of what the current state is. See [https://issues.k8s.io/34274](https://issues.k8s.io/34274).

{{ header }} Syntax

```shell
werf kubectl apply (-f FILENAME | -k DIRECTORY) [options]
```

{{ header }} Examples

```shell
  # Apply the configuration in pod.json to a pod
  kubectl apply -f ./pod.json
  
  # Apply resources from a directory containing kustomization.yaml - e.g. dir/kustomization.yaml
  kubectl apply -k dir/
  
  # Apply the JSON passed into stdin to a pod
  cat pod.json | kubectl apply -f -
  
  # Note: --prune is still in Alpha
  # Apply the configuration in manifest.yaml that matches label app=nginx and delete all other resources that are not in the file and match label app=nginx
  kubectl apply --prune -f manifest.yaml -l app=nginx
  
  # Apply the configuration in manifest.yaml and delete all the other config maps that are not in the file
  kubectl apply --prune -f manifest.yaml --all --prune-whitelist=core/v1/ConfigMap
```

{{ header }} Options

```shell
      --all=false
            Select all resources in the namespace of the specified resource types.
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
      --field-manager='kubectl-client-side-apply'
            Name of the manager used to track field ownership.
  -f, --filename=[]
            that contains the configuration to apply
      --force=false
            If true, immediately remove resources from API and bypass graceful deletion. Note that  
            immediate deletion of some resources may result in inconsistency or data loss and       
            requires confirmation.
      --force-conflicts=false
            If true, server-side apply will force the changes against conflicts.
      --grace-period=-1
            Period of time in seconds given to the resource to terminate gracefully. Ignored if     
            negative. Set to 1 for immediate shutdown. Can only be set to 0 when --force is true    
            (force deletion).
  -k, --kustomize=''
            Process a kustomization directory. This flag can`t be used together with -f or -R.
      --openapi-patch=true
            If true, use openapi to calculate diff when the openapi presents and the resource can   
            be found in the openapi spec. Otherwise, fall back to use baked-in types.
  -o, --output=''
            Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile
            |jsonpath|jsonpath-as-json|jsonpath-file.
      --overwrite=true
            Automatically resolve conflicts between the modified and live configuration by using    
            values from the modified configuration
      --prune=false
            Automatically delete resource objects, that do not appear in the configs and are        
            created by either apply or create --save-config. Should be used with either -l or --all.
      --prune-whitelist=[]
            Overwrite the default whitelist with <group/version/kind> for --prune
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2)
      --server-side=false
            If true, apply runs in the server instead of the client.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
      --timeout=0s
            The length of time to wait before giving up on a delete, zero means determine a timeout 
            from the size of the object
      --validate=true
            If true, use a schema to validate the input before sending it
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
            The name of the kubeconfig context to use
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure
      --kubeconfig=''
            Path to the kubeconfig file to use for CLI requests.
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
      --token=''
            Bearer token for authentication to the API server
      --user=''
            The name of the kubeconfig user to use
      --username=''
            Username for basic authentication to the API server
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

