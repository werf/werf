{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Update environment variables on a pod template.

List environment variable definitions in one or more pods, pod templates. Add, update, or remove container environment variable definitions in one or more pod templates (within replication controllers or deployment configurations). View or modify the environment variable definitions on all containers in the specified pods or pod templates, or just those that match a wildcard.

If `--env -` is passed, environment variables can be read from STDIN using the standard env syntax.

Possible resources include (case insensitive):
* pod (po),
* replicationcontroller (rc),
* deployment (deploy),
* daemonset (ds),
* statefulset (sts),
* cronjob (cj),
 replicaset (rs).

{{ header }} Syntax

```shell
werf kubectl set env RESOURCE/NAME KEY_1=VAL_1 ... KEY_N=VAL_N [options]
```

{{ header }} Examples

```shell
  # Update deployment 'registry' with a new environment variable
  kubectl set env deployment/registry STORAGE_DIR=/local
  
  # List the environment variables defined on a deployments 'sample-build'
  kubectl set env deployment/sample-build --list
  
  # List the environment variables defined on all pods
  kubectl set env pods --all --list
  
  # Output modified deployment in YAML, and does not alter the object on the server
  kubectl set env deployment/sample-build STORAGE_DIR=/data -o yaml
  
  # Update all containers in all replication controllers in the project to have ENV=prod
  kubectl set env rc --all ENV=prod
  
  # Import environment from a secret
  kubectl set env --from=secret/mysecret deployment/myapp
  
  # Import environment from a config map with a prefix
  kubectl set env --from=configmap/myconfigmap --prefix=MYSQL_ deployment/myapp
  
  # Import specific keys from a config map
  kubectl set env --keys=my-example-key --from=configmap/myconfigmap deployment/myapp
  
  # Remove the environment variable ENV from container 'c1' in all deployment configs
  kubectl set env deployments --all --containers="c1" ENV-
  
  # Remove the environment variable ENV from a deployment definition on disk and
  # update the deployment config on the server
  kubectl set env -f deploy.json ENV-
  
  # Set some of the local shell environment into a deployment config on the server
  env | grep RAILS_ | kubectl set env -e - deployment/registry
```

{{ header }} Options

```shell
      --all=false
            If true, select all resources in the namespace of the specified resource types
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
  -c, --containers='*'
            The names of containers in the selected pod templates to change - may use wildcards
      --dry-run='none'
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
  -e, --env=[]
            Specify a key-value pair for an environment variable to set into each container.
      --field-manager='kubectl-set'
            Name of the manager used to track field ownership.
  -f, --filename=[]
            Filename, directory, or URL to files the resource to update the env
      --from=''
            The name of a resource from which to inject environment variables
      --keys=[]
            Comma-separated list of keys to import from specified resource
  -k, --kustomize=''
            Process the kustomization directory. This flag can`t be used together with -f or -R.
      --list=false
            If true, display the environment and any changes in the standard format. this flag will 
            removed when we have kubectl view env.
      --local=false
            If true, set env will NOT contact api-server but run locally.
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --overwrite=true
            If true, allow environment to be overwritten, otherwise reject updates that overwrite   
            existing environment.
      --prefix=''
            Prefix to append to variable names
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
      --resolve=false
            If true, show secret or configmap references when listing variables
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2). Matching objects must satisfy all of the specified label      
            constraints.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
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

