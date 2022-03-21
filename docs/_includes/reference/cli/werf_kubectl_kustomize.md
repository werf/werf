{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build a set of KRM resources using a &#39;kustomization.yaml&#39; file. The DIR argument must be a path to a directory containing &#39;kustomization.yaml&#39;, or a git repository URL with a path suffix specifying same with respect to the repository root. If DIR is omitted, &#39;.&#39; is assumed.

{{ header }} Syntax

```shell
werf kubectl kustomize DIR [flags] [options]
```

{{ header }} Examples

```shell
  # Build the current working directory
  kubectl kustomize
  
  # Build some shared configuration directory
  kubectl kustomize /home/config/production
  
  # Build from github
  kubectl kustomize https://github.com/kubernetes-sigs/kustomize.git/examples/helloWorld?ref=v1.0.6
```

{{ header }} Options

```shell
      --as-current-user=false
            use the uid and gid of the command executor to run the function in the container
      --enable-alpha-plugins=false
            enable kustomize plugins
      --enable-helm=false
            Enable use of the Helm chart inflator generator.
      --enable-managedby-label=false
            enable adding app.kubernetes.io/managed-by
  -e, --env=[]
            a list of environment variables to be used by functions
      --helm-command='helm'
            helm command (path to executable)
      --load-restrictor='LoadRestrictionsRootOnly'
            if set to `LoadRestrictionsNone`, local kustomizations may load files from outside      
            their root. This does, however, break the relocatability of the kustomization.
      --mount=[]
            a list of storage options read from the filesystem
      --network=false
            enable network access for functions that declare it
      --network-name='bridge'
            the docker network to run the container in
  -o, --output=''
            If specified, write output to this path.
      --reorder='legacy'
            Reorder the resources just before output. Use `legacy` to apply a legacy reordering     
            (Namespaces first, Webhooks last, etc). Use `none` to suppress a final reordering.
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

