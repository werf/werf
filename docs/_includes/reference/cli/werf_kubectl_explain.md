{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
List the fields for supported resources.

This command describes the fields associated with each supported API resource. Fields are identified via a simple JSON Path identifier:
```
<type>.<fieldName>[.<fieldName>]
```
Add the `--recursive` flag to display all of the fields at once without descriptions. Information about each field is retrieved from the server in OpenAPI format.

Use `kubectl api-resources` for a complete list of supported resources.

{{ header }} Syntax

```shell
werf kubectl explain TYPE [--recursive=FALSE|TRUE] [--api-version=api-version-group] [--output=plaintext|plaintext-openapiv2] [options]
```

{{ header }} Examples

```shell
  # Get the documentation of the resource and its fields
  kubectl explain pods
  
  # Get all the fields in the resource
  kubectl explain pods --recursive
  
  # Get the explanation for deployment in supported api versions
  kubectl explain deployments --api-version=apps/v1
  
  # Get the documentation of a specific field of a resource
  kubectl explain pods.spec.containers
  
  # Get the documentation of resources in different format
  kubectl explain deployment --output=plaintext-openapiv2
```

{{ header }} Options

```shell
      --api-version=''
            Use given api-version (group/version) of the resource.
      --output='plaintext'
            Format in which to render the schema. Valid values are: (plaintext,                     
            plaintext-openapiv2).
      --recursive=false
            When true, print the name of all the fields recursively. Otherwise, print the available 
            fields with their description.
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

