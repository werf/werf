{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Set kuberc configuration values.

{{ header }} Syntax

```shell
werf kubectl alpha kuberc set --section (defaults|aliases) --command COMMAND [options]
```

{{ header }} Examples

```shell
  # Set default output format for 'get' command
  kubectl alpha kuberc set --section defaults --command get --option output=wide
  
  # Set default output format for a subcommand
  kubectl alpha kuberc set --section defaults --command "set env" --option output=yaml
  
  # Create an alias 'getn' for 'get' command with prepended 'nodes' resource
  kubectl alpha kuberc set --section aliases --name getn --command get --prependarg nodes --option output=wide
  
  # Create an alias 'runx' for 'run' command with appended arguments
  kubectl alpha kuberc set --section aliases --name runx --command run --option image=nginx --appendarg "--" --appendarg custom-arg1
  
  # Overwrite an existing default
  kubectl alpha kuberc set --section defaults --command get --option output=json --overwrite
```

{{ header }} Options

```shell
      --appendarg=[]
            Argument to append to the command (can be specified multiple times, for aliases only)
      --command=""
            Command to configure (e.g., `get`, `create`, `set env`)
      --kuberc=""
            Path to the kuberc file to use for preferences. This can be disabled by exporting       
            KUBECTL_KUBERC=false feature gate or turning off the feature KUBERC=off.
      --name=""
            Alias name (required for --section=aliases)
      --option=[]
            Flag option in the form flag=value (can be specified multiple times)
      --overwrite=false
            Allow overwriting existing entries
      --prependarg=[]
            Argument to prepend to the command (can be specified multiple times, for aliases only)
      --section=""
            Section to modify: `defaults` or `aliases`
```

{{ header }} Options inherited from parent commands

```shell
      --as=""
            Username to impersonate for the operation. User could be a regular user or a service    
            account in a namespace.
      --as-group=[]
            Group to impersonate for the operation, this flag can be repeated to specify multiple   
            groups.
      --as-uid=""
            UID to impersonate for the operation.
      --as-user-extra=[]
            User extras to impersonate for the operation, this flag can be repeated to specify      
            multiple values for the same key.
      --cache-dir="~/.kube/cache"
            Default cache directory
      --certificate-authority=""
            Path to a cert file for the certificate authority
      --client-certificate=""
            Path to a client certificate file for TLS
      --client-key=""
            Path to a client key file for TLS
      --cluster=""
            The name of the kubeconfig cluster to use
      --context=""
            The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)
      --disable-compression=false
            If true, opt-out of response compression for all requests to the server
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --kube-config-base64=""
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kubeconfig=""
            Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or      
            $WERF_KUBECONFIG, or $KUBECONFIG). Ignored if kubeconfig passed as base64.
      --log-flush-frequency=5s
            Maximum number of seconds between log flushes
      --match-server-version=false
            Require server version to match client version
  -n, --namespace=""
            If present, the namespace scope for this CLI request
      --password=""
            Password for basic authentication to the API server
      --profile="none"
            Name of profile to capture. One of                                                      
            (none|cpu|heap|goroutine|threadcreate|block|mutex|trace)
      --profile-output="profile.pprof"
            Name of the file to write the profile to
      --request-timeout="0"
            The length of time to wait before giving up on a single server request. Non-zero values 
            should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don`t 
            timeout requests.
  -s, --server=""
            The address and port of the Kubernetes API server
      --tls-server-name=""
            Server name to use for server certificate validation. If it is not provided, the        
            hostname used to contact the server is used
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --token=""
            Bearer token for authentication to the API server
      --user=""
            The name of the kubeconfig user to use
      --username=""
            Username for basic authentication to the API server
  -v, --v=0
            number for the log level verbosity
      --vmodule=
            comma-separated list of pattern=N settings for file-filtered logging (only works for    
            the default text log format)
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

