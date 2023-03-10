{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Attach to a process that is already running inside an existing container.

{{ header }} Syntax

```shell
werf kubectl attach (POD | TYPE/NAME) -c CONTAINER [options]
```

{{ header }} Examples

```shell
  # Get output from running pod mypod; use the 'kubectl.kubernetes.io/default-container' annotation
  # for selecting the container to be attached or the first container in the pod will be chosen
  kubectl attach mypod
  
  # Get output from ruby-container from pod mypod
  kubectl attach mypod -c ruby-container
  
  # Switch to raw terminal mode; sends stdin to 'bash' in ruby-container from pod mypod
  # and sends stdout/stderr from 'bash' back to the client
  kubectl attach mypod -c ruby-container -i -t
  
  # Get output from the first pod of a replica set named nginx
  kubectl attach rs/nginx
```

{{ header }} Options

```shell
  -c, --container=''
            Container name. If omitted, use the kubectl.kubernetes.io/default-container annotation  
            for selecting the container to be attached or the first container in the pod will be    
            chosen
      --pod-running-timeout=1m0s
            The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one    
            pod is running
  -q, --quiet=false
            Only print output from the remote session
  -i, --stdin=false
            Pass stdin to the container
  -t, --tty=false
            Stdin is a TTY
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

