{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Copy files and directories to and from containers.

{{ header }} Syntax

```shell
werf kubectl cp <file-spec-src> <file-spec-dest> [options]
```

{{ header }} Examples

```shell
  # !!!Important Note!!!
  # Requires that the 'tar' binary is present in your container
  # image.  If 'tar' is not present, 'kubectl cp' will fail.
  #
  # For advanced use cases, such as symlinks, wildcard expansion or
  # file mode preservation, consider using 'kubectl exec'.
  
  # Copy /tmp/foo local file to /tmp/bar in a remote pod in namespace <some-namespace>
  tar cf - /tmp/foo | kubectl exec -i -n <some-namespace> <some-pod> -- tar xf - -C /tmp/bar
  
  # Copy /tmp/foo from a remote pod to /tmp/bar locally
  kubectl exec -n <some-namespace> <some-pod> -- tar cf - /tmp/foo | tar xf - -C /tmp/bar
  
  # Copy /tmp/foo_dir local directory to /tmp/bar_dir in a remote pod in the default namespace
  kubectl cp /tmp/foo_dir <some-pod>:/tmp/bar_dir
  
  # Copy /tmp/foo local file to /tmp/bar in a remote pod in a specific container
  kubectl cp /tmp/foo <some-pod>:/tmp/bar -c <specific-container>
  
  # Copy /tmp/foo local file to /tmp/bar in a remote pod in namespace <some-namespace>
  kubectl cp /tmp/foo <some-namespace>/<some-pod>:/tmp/bar
  
  # Copy /tmp/foo from a remote pod to /tmp/bar locally
  kubectl cp <some-namespace>/<some-pod>:/tmp/foo /tmp/bar
```

{{ header }} Options

```shell
  -c, --container=''
            Container name. If omitted, use the kubectl.kubernetes.io/default-container annotation  
            for selecting the container to be attached or the first container in the pod will be    
            chosen
      --no-preserve=false
            The copied file/directory`s ownership and permissions will not be preserved in the      
            container
      --retries=0
            Set number of retries to complete a copy operation from a container. Specify 0 to       
            disable or any negative value for infinite retrying. The default is 0 (no retry).
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

