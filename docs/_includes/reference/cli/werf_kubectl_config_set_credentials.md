{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Set a user entry in kubeconfig.

Specifying a name that already exists will merge new fields on top of existing values:
* Client-certificate flags: `--client-certificate=certfile`, `--client-key=keyfile`;
* Bearer token flags: `--token=bearer_token`;
* Basic auth flags: `--username=basic_user`, `--password=basic_password`.

Bearer token and basic auth are mutually exclusive.

{{ header }} Syntax

```shell
werf kubectl config set-credentials NAME [--client-certificate=path/to/certfile] [--client-key=path/to/keyfile] [--token=bearer_token] [--username=basic_user] [--password=basic_password] [--auth-provider=provider_name] [--auth-provider-arg=key=value] [--exec-command=exec_command] [--exec-api-version=exec_api_version] [--exec-arg=arg] [--exec-env=key=value] [options]
```

{{ header }} Examples

```shell
  # Set only the "client-key" field on the "cluster-admin"
  # entry, without touching other values
  kubectl config set-credentials cluster-admin --client-key=~/.kube/admin.key
  
  # Set basic auth for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --username=admin --password=uXFGweU9l35qcif
  
  # Embed client certificate data in the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --client-certificate=~/.kube/admin.crt --embed-certs=true
  
  # Enable the Google Compute Platform auth provider for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --auth-provider=gcp
  
  # Enable the OpenID Connect auth provider for the "cluster-admin" entry with additional arguments
  kubectl config set-credentials cluster-admin --auth-provider=oidc --auth-provider-arg=client-id=foo --auth-provider-arg=client-secret=bar
  
  # Remove the "client-secret" config value for the OpenID Connect auth provider for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --auth-provider=oidc --auth-provider-arg=client-secret-
  
  # Enable new exec auth plugin for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-command=/path/to/the/executable --exec-api-version=client.authentication.k8s.io/v1beta1
  
  # Define new exec auth plugin arguments for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-arg=arg1 --exec-arg=arg2
  
  # Create or update exec auth plugin environment variables for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-env=key1=val1 --exec-env=key2=val2
  
  # Remove exec auth plugin environment variables for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-env=var-to-remove-
```

{{ header }} Options

```shell
      --auth-provider=''
            Auth provider for the user entry in kubeconfig
      --auth-provider-arg=[]
            `key=value` arguments for the auth provider
      --client-certificate=''
            Path to client-certificate file for the user entry in kubeconfig
      --client-key=''
            Path to client-key file for the user entry in kubeconfig
      --embed-certs=false
            Embed client cert/key for the user entry in kubeconfig
      --exec-api-version=''
            API version of the exec credential plugin for the user entry in kubeconfig
      --exec-arg=[]
            New arguments for the exec credential plugin command for the user entry in kubeconfig
      --exec-command=''
            Command for the exec credential plugin for the user entry in kubeconfig
      --exec-env=[]
            `key=value` environment values for the exec credential plugin
      --password=''
            password for the user entry in kubeconfig
      --token=''
            token for the user entry in kubeconfig
      --username=''
            username for the user entry in kubeconfig
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
            use a particular kubeconfig file
      --match-server-version=false
            Require server version to match client version
  -n, --namespace=''
            If present, the namespace scope for this CLI request
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
      --user=''
            The name of the kubeconfig user to use
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

