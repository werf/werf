{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Create a role binding for a particular role or cluster role.

{{ header }} Syntax

```shell
werf kubectl create rolebinding NAME --clusterrole=NAME|--role=NAME [--user=username] [--group=groupname] [--serviceaccount=namespace:serviceaccountname] [--dry-run=server|client|none] [options]
```

{{ header }} Examples

```shell
  # Create a role binding for user1, user2, and group1 using the admin cluster role
  kubectl create rolebinding admin --clusterrole=admin --user=user1 --user=user2 --group=group1
  
  # Create a role binding for serviceaccount monitoring:sa-dev using the admin role
  kubectl create rolebinding admin-binding --role=admin --serviceaccount=monitoring:sa-dev
```

{{ header }} Options

```shell
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --clusterrole=""
            ClusterRole this RoleBinding should reference
      --dry-run="none"
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
      --field-manager="kubectl-create"
            Name of the manager used to track field ownership.
      --group=[]
            Groups to bind to the role. The flag can be repeated to add multiple groups.
  -o, --output=""
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --role=""
            Role this RoleBinding should reference
      --save-config=false
            If true, the configuration of current object will be saved in its annotation.           
            Otherwise, the annotation will be unchanged. This flag is useful when you want to       
            perform kubectl apply on this object in the future.
      --serviceaccount=[]
            Service accounts to bind to the role, in the format <namespace>:<name>. The flag can be 
            repeated to add multiple service accounts.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=""
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
      --user=[]
            Usernames to bind to the role. The flag can be repeated to add multiple users.
      --validate="strict"
            Must be one of: strict (or true), warn, ignore (or false).
            		"true" or "strict" will use a schema to validate the input and fail the request if    
            invalid. It will perform server side validation if ServerSideFieldValidation is enabled 
            on the api-server, but will fall back to less reliable client-side validation if not.
            		"warn" will warn about unknown or duplicate fields without blocking the request if    
            server-side field validation is enabled on the API server, and behave as "ignore"       
            otherwise.
            		"false" or "ignore" will not perform any schema validation, silently dropping any     
            unknown or duplicate fields.
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
            Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)
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

