{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Creates a proxy server or application-level gateway between localhost and the Kubernetes API server. It also allows serving static content over specified HTTP path. All incoming data enters through one port and gets forwarded to the remote Kubernetes API server port, except for the path matching the static content path.

{{ header }} Syntax

```shell
werf kubectl proxy [--port=PORT] [--www=static-dir] [--www-prefix=prefix] [--api-prefix=prefix] [options]
```

{{ header }} Examples

```shell
  # To proxy all of the Kubernetes API and nothing else
  kubectl proxy --api-prefix=/
  
  # To proxy only part of the Kubernetes API and also some static files
  # You can get pods info with 'curl localhost:8001/api/v1/pods'
  kubectl proxy --www=/my/files --www-prefix=/static/ --api-prefix=/api/
  
  # To proxy the entire Kubernetes API at a different root
  # You can get pods info with 'curl localhost:8001/custom/api/v1/pods'
  kubectl proxy --api-prefix=/custom/
  
  # Run a proxy to the Kubernetes API server on port 8011, serving static content from ./local/www/
  kubectl proxy --port=8011 --www=./local/www/
  
  # Run a proxy to the Kubernetes API server on an arbitrary local port
  # The chosen port for the server will be output to stdout
  kubectl proxy --port=0
  
  # Run a proxy to the Kubernetes API server, changing the API prefix to k8s-api
  # This makes e.g. the pods API available at localhost:8001/k8s-api/v1/pods/
  kubectl proxy --api-prefix=/k8s-api
```

{{ header }} Options

```shell
      --accept-hosts='^localhost$,^127\.0\.0\.1$,^\[::1\]$'
            Regular expression for hosts that the proxy should accept.
      --accept-paths='^.*'
            Regular expression for paths that the proxy should accept.
      --address='127.0.0.1'
            The IP address on which to serve on.
      --api-prefix='/'
            Prefix to serve the proxied API under.
      --append-server-path=false
            If true, enables automatic path appending of the kube context server path to each       
            request.
      --disable-filter=false
            If true, disable request filtering in the proxy. This is dangerous, and can leave you   
            vulnerable to XSRF attacks, when used with an accessible port.
      --keepalive=0s
            keepalive specifies the keep-alive period for an active network connection. Set to 0 to 
            disable keepalive.
  -p, --port=8001
            The port on which to run the proxy. Set to 0 to pick a random port.
      --reject-methods='^$'
            Regular expression for HTTP methods that the proxy should reject (example               
            --reject-methods=`POST,PUT,PATCH`). 
      --reject-paths='^/api/.*/pods/.*/exec,^/api/.*/pods/.*/attach'
            Regular expression for paths that the proxy should reject. Paths specified here will be 
            rejected even accepted by --accept-paths.
  -u, --unix-socket=''
            Unix socket on which to run the proxy.
  -w, --www=''
            Also serve static files from the given directory under the specified prefix.
  -P, --www-prefix='/static/'
            Prefix to serve static files under, if static file directory is specified.
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

