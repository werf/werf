{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Set an individual value in a kubeconfig file.
* `PROPERTY_NAME` is a dot delimited name where each token represents either an attribute name or a map key.  Map keys may not contain dots.
* `PROPERTY_VALUE` is the new value you want to set. Binary fields such as `certificate-authority-data` expect a base64 encoded string unless the `--set-raw-bytes` flag is used.

Specifying an attribute name that already exists will merge new fields on top of existing values.

{{ header }} Syntax

```shell
werf kubectl config set PROPERTY_NAME PROPERTY_VALUE [options]
```

{{ header }} Examples

```shell
  # Set the server field on the my-cluster cluster to https://1.2.3.4
  kubectl config set clusters.my-cluster.server https://1.2.3.4
  
  # Set the certificate-authority-data field on the my-cluster cluster
  kubectl config set clusters.my-cluster.certificate-authority-data $(echo "cert_data_here" | base64 -i -)
  
  # Set the cluster field in the my-context context to my-cluster
  kubectl config set contexts.my-context.cluster my-cluster
  
  # Set the client-key-data field in the cluster-admin user using --set-raw-bytes option
  kubectl config set users.cluster-admin.client-key-data cert_data_here --set-raw-bytes=true
```

{{ header }} Options

```shell
      --set-raw-bytes=false
            When writing a []byte PROPERTY_VALUE, write the given string directly without base64    
            decoding.
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
            use a particular kubeconfig file
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

