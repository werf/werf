{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Render chart templates locally and display the output.

Any values that would normally be looked up or retrieved in-cluster will be
faked locally. Additionally, none of the server-side testing of chart validity
(e.g. whether an API is supported) is done.


{{ header }} Syntax

```shell
werf helm template [NAME] [CHART] [flags] [options]
```

{{ header }} Options

```shell
      --add-annotation=[]
            Add annotation to deploying resources (can specify multiple).
            Format: annoName=annoValue.
            Also, can be specified with $WERF_ADD_ANNOTATION_* (e.g.                                
            $WERF_ADD_ANNOTATION_1=annoName1=annoValue1,                                            
            $WERF_ADD_ANNOTATION_2=annoName2=annoValue2)
      --add-label=[]
            Add label to deploying resources (can specify multiple).
            Format: labelName=labelValue.
            Also, can be specified with $WERF_ADD_LABEL_* (e.g.                                     
            $WERF_ADD_LABEL_1=labelName1=labelValue1, $WERF_ADD_LABEL_2=labelName2=labelValue2)
  -a, --api-versions=[]
            Kubernetes api versions used for Capabilities.APIVersions
      --atomic=false
            if set, the installation process deletes the installation on failure. The --wait flag   
            will be set automatically if --atomic is used
      --ca-file=''
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=''
            identify HTTPS client using this SSL certificate file
      --cleanup-on-fail=false
            allow deletion of new resources created in this installation when install fails
      --create-namespace=false
            create the release namespace if not present
      --dependency-update=false
            update dependencies if they are missing before installing the chart
      --deploy-report-path=''
            save deploy report in JSON to the specified path
      --description=''
            add a custom description
      --devel=false
            use development versions, too. Equivalent to version `>0.0.0-0`. If --version is set,   
            this is ignored
      --disable-openapi-validation=false
            if set, the installation process will not validate rendered templates against the       
            Kubernetes OpenAPI Schema
      --dry-run=''
            simulate an install. If --dry-run is set with no option being specified or as           
            `--dry-run=client`, it will not attempt cluster connections. Setting `--dry-run=server` 
            allows attempting cluster connections.
      --enable-dns=false
            enable DNS lookups when rendering templates
      --force=false
            force resource updates through a replacement strategy
  -g, --generate-name=false
            generate the name (and omit the NAME parameter)
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --include-crds=false
            include CRDs in the templated output
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the chart download
      --is-upgrade=false
            set .Release.IsUpgrade instead of .Release.IsInstall
      --key-file=''
            identify HTTPS client using this SSL key file
      --keyring='~/.gnupg/pubring.gpg'
            location of public keys used for verification
      --kube-version=''
            Kubernetes version used for Capabilities.KubeVersion
  -l, --labels=[]
            Labels that would be added to release metadata. Should be divided by comma.
      --name-template=''
            specify template used to name the release
      --no-hooks=false
            prevent hooks from running during install
      --output-dir=''
            writes the executed templates to files in output-dir instead of stdout
      --pass-credentials=false
            pass credentials to all domains
      --password=''
            chart repository password where to locate the requested chart
      --plain-http=false
            use insecure HTTP connections for the chart download
      --post-renderer=
            the path to an executable to be used for post rendering. If it exists in $PATH, the     
            binary will be used, otherwise it will try to look for the executable at the given path
      --post-renderer-args=[]
            an argument to the post-renderer (can specify multiple)
      --release-name=false
            use release name in the output-dir path.
      --render-subchart-notes=false
            if set, render subchart notes along with the parent
      --replace=false
            re-use the given name, only if that name is a deleted release which remains in the      
            history. This is unsafe in production
      --repo=''
            chart repository url where to locate the requested chart
      --secret-values=[]
            Specify helm secret values in a YAML file (can specify multiple).
            Also, can be defined with $WERF_SECRET_VALUES_* (e.g.                                   
            $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml,                                  
            $WERF_SECRET_VALUES_DB=.helm/secret_values_db.yaml)
      --set=[]
            set values on the command line (can specify multiple or separate values with commas:    
            key1=val1,key2=val2)
      --set-file=[]
            set values from respective files specified via the command line (can specify multiple   
            or separate values with commas: key1=path1,key2=path2)
      --set-json=[]
            set JSON values on the command line (can specify multiple or separate values with       
            commas: key1=jsonval1,key2=jsonval2)
      --set-literal=[]
            set a literal STRING value on the command line
      --set-string=[]
            set STRING values on the command line (can specify multiple or separate values with     
            commas: key1=val1,key2=val2)
  -s, --show-only=[]
            only show manifests rendered from the given templates
      --skip-crds=false
            if set, no CRDs will be installed. By default, CRDs are installed if not already present
      --skip-tests=false
            skip tests from templated output
      --timeout=5m0s
            time to wait for any individual Kubernetes operation (like Jobs for hooks)
      --username=''
            chart repository username where to locate the requested chart
      --validate=false
            validate your manifests against the Kubernetes cluster you are currently pointing at.   
            This is the same validation performed on an install
  -f, --values=[]
            specify values in a YAML file or a URL (can specify multiple)
      --verify=false
            verify the package before using it
      --version=''
            specify a version constraint for the chart version to use. This constraint can be a     
            specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is   
            not specified, the latest version is used
      --wait=false
            if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a       
            Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release  
            as successful. It will wait for as long as --timeout
      --wait-for-jobs=false
            if set and --wait enabled, will wait until all Jobs have been completed before marking  
            the release as successful. It will wait for as long as --timeout
```

{{ header }} Options inherited from parent commands

```shell
      --hooks-status-progress-period=5
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
            $KUBECONFIG)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=''
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-color-mode='auto'
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format='2006-01-02T15:04:05Z07:00'
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
  -n, --namespace=''
            namespace scope for this request
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

