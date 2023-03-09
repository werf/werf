{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Edit the latest last-applied-configuration annotations of resources from the default editor.

The `edit-last-applied` command allows you to directly edit any API resource you can retrieve via the command-line tools. It will open the editor defined by your `KUBE_EDITOR`, or `EDITOR` environment variables, or fall back to `vi` for Linux or `notepad` for Windows. You can edit multiple objects, although changes are applied one at a time. The command accepts file names as well as command-line arguments, although the files you point to must be previously saved versions of resources.

The default format is YAML. To edit in JSON, specify `-o json`.

The flag `--windows-line-endings` can be used to force Windows line endings, otherwise the default for your operating system will be used.

In the event an error occurs while updating, a temporary file will be created on disk that contains your unapplied changes. The most common error when updating a resource is another editor changing the resource on the server. When this occurs, you will have to apply your changes to the newer version of the resource, or update your temporary saved copy to include the latest resource version.

{{ header }} Syntax

```shell
werf kubectl apply edit-last-applied (RESOURCE/NAME | -f FILENAME) [options]
```

{{ header }} Examples

```shell
  # Edit the last-applied-configuration annotations by type/name in YAML
  kubectl apply edit-last-applied deployment/nginx
  
  # Edit the last-applied-configuration annotations by file in JSON
  kubectl apply edit-last-applied -f deploy.yaml -o json
```

{{ header }} Options

```shell
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --field-manager='kubectl-client-side-apply'
            Name of the manager used to track field ownership.
  -f, --filename=[]
            Filename, directory, or URL to files to use to edit the resource
  -k, --kustomize=''
            Process the kustomization directory. This flag can`t be used together with -f or -R.
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
  -R, --recursive=false
            Process the directory used in -f, --filename recursively. Useful when you want to       
            manage related manifests organized within the same directory.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
      --validate='strict'
            Must be one of: strict (or true), warn, ignore (or false).
            		"true" or "strict" will use a schema to validate the input and fail the request if    
            invalid. It will perform server side validation if ServerSideFieldValidation is enabled 
            on the api-server, but will fall back to less reliable client-side validation if not.
            		"warn" will warn about unknown or duplicate fields without blocking the request if    
            server-side field validation is enabled on the API server, and behave as "ignore"       
            otherwise.
            		"false" or "ignore" will not perform any schema validation, silently dropping any     
            unknown or duplicate fields.
      --windows-line-endings=false
            Defaults to the line ending native to your platform.
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

