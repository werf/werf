{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Show the status of the rollout.

 By default &#39;rollout status&#39; will watch the status of the latest rollout until it&#39;s done. If you don&#39;t want to wait for the rollout to finish then you can use --watch=false. Note that if a new rollout starts in-between, then &#39;rollout status&#39; will continue watching the latest revision. If you want to pin to a specific revision and abort if it is rolled over by another revision, use --revision=N where N is the revision you need to watch for.

{{ header }} Syntax

```shell
werf kubectl rollout status (TYPE NAME | TYPE/NAME) [flags] [options]
```

{{ header }} Examples

```shell
  # Watch the rollout status of a deployment
  kubectl rollout status deployment/nginx
```

{{ header }} Options

```shell
  -f, --filename=[]
            Filename, directory, or URL to files identifying the resource to get from a server.
  -k, --kustomize=''
            Process the kustomization directory. This flag can`t be used together with -f or -R.
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
      --revision=0
            Pin to a specific revision for showing its status. Defaults to 0 (last revision).
      --timeout=0s
            The length of time to wait before ending watch, zero means never. Any other values      
            should contain a corresponding time unit (e.g. 1s, 2m, 3h).
  -w, --watch=true
            Watch the status of the rollout until it`s done.
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

