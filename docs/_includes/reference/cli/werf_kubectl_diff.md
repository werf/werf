{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Diff configurations specified by file name or stdin between the current online configuration, and the configuration as it would be if applied.

 The output is always YAML.

 KUBECTL_EXTERNAL_DIFF environment variable can be used to select your own diff command. Users can use external commands with params too, example: KUBECTL_EXTERNAL_DIFF=&#34;colordiff -N -u&#34;

 By default, the &#34;diff&#34; command available in your path will be run with the &#34;-u&#34; (unified diff) and &#34;-N&#34; (treat absent files as empty) options.

 Exit status: 0 No differences were found. 1 Differences were found. &gt;1 Kubectl or diff failed with an error.

 Note: KUBECTL_EXTERNAL_DIFF, if used, is expected to follow that convention.

{{ header }} Syntax

```shell
werf kubectl diff -f FILENAME [options]
```

{{ header }} Examples

```shell
  # Diff resources included in pod.json
  kubectl diff -f pod.json
  
  # Diff file read from stdin
  cat service.yaml | kubectl diff -f -
```

{{ header }} Options

```shell
      --field-manager='kubectl-client-side-apply'
            Name of the manager used to track field ownership.
  -f, --filename=[]
            Filename, directory, or URL to files contains the configuration to diff
      --force-conflicts=false
            If true, server-side apply will force the changes against conflicts.
  -k, --kustomize=''
            Process the kustomization directory. This flag can`t be used together with -f or -R.
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2)
      --server-side=false
            If true, apply runs in the server instead of the client.
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

