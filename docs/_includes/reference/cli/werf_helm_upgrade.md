{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
This command upgrades a release to a new version of a chart.

The upgrade arguments must be a release and chart. The chart argument can be either: a chart reference(`example/mariadb`), a path to a chart directory, a packaged chart, or a fully qualified URL. For chart references, the latest version will be specified unless the `--version` flag is set.

To override values in a chart, use either the `--values` flag and pass in a file or use the `--set` flag and pass configuration from the command line, to force string values, use `--set-string`. You can use `--set-file` to set individual values from a file when the value itself is too long for the command line or is dynamically generated.

You can specify the `--values`/`-f` flag multiple times. The priority will be given to the last (right-most) file specified. For example, if both `myvalues.yaml` and `override.yaml` contained a key called `Test`, the value set in `override.yaml` would take precedence:
```
$ helm upgrade -f myvalues.yaml -f override.yaml redis ./redis
```
You can specify the `--set` flag multiple times. The priority will be given to the last (right-most) set specified. For example, if both `bar` and `newbar` values are set for a key called `foo`, the `newbar` value would take precedence:
```
$ helm upgrade --set foo=bar --set foo=newbar redis ./redis
```


{{ header }} Syntax

```shell
werf helm upgrade [RELEASE] [CHART] [flags] [options]
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
      --atomic=false
            if set, upgrade process rolls back changes made in case of failed upgrade. The --wait   
            flag will be set automatically if --atomic is used
      --ca-file=''
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=''
            identify HTTPS client using this SSL certificate file
      --cleanup-on-fail=false
            allow deletion of new resources created in this upgrade when upgrade fails
      --create-namespace=false
            if --install is set, create the release namespace if not present
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
            if set, the upgrade process will not validate rendered templates against the Kubernetes 
            OpenAPI Schema
      --dry-run=''
            simulate an install. If --dry-run is set with no option being specified or as           
            `--dry-run=client`, it will not attempt cluster connections. Setting `--dry-run=server` 
            allows attempting cluster connections.
      --enable-dns=false
            enable DNS lookups when rendering templates
      --force=false
            force resource updates through a replacement strategy
      --history-max=10
            limit the maximum number of revisions saved per release. Use 0 for no limit
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the chart download
  -i, --install=false
            if a release by this name doesn`t already exist, run an install
      --key-file=''
            identify HTTPS client using this SSL key file
      --keyring='~/.gnupg/pubring.gpg'
            location of public keys used for verification
  -l, --labels=[]
            Labels that would be added to release metadata. Should be separated by comma. Original  
            release labels will be merged with upgrade labels. You can unset label using null.
      --no-hooks=false
            disable pre/post upgrade hooks
  -o, --output=table
            prints the output in the specified format. Allowed values: table, json, yaml
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
      --render-subchart-notes=false
            if set, render subchart notes along with the parent
      --repo=''
            chart repository url where to locate the requested chart
      --reset-then-reuse-values=false
            when upgrading, reset the values to the ones built into the chart, apply the last       
            release`s values and merge in any overrides from the command line via --set and -f. If  
            `--reset-values` or `--reuse-values` is specified, this is ignored
      --reset-values=false
            when upgrading, reset the values to the ones built into the chart
      --reuse-values=false
            when upgrading, reuse the last release`s values and merge in any overrides from the     
            command line via --set and -f. If `--reset-values` is specified, this is ignored
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
      --skip-crds=false
            if set, no CRDs will be installed when an upgrade is performed with install flag        
            enabled. By default, CRDs are installed if not already present, when an upgrade is      
            performed with install flag enabled
      --timeout=5m0s
            time to wait for any individual Kubernetes operation (like Jobs for hooks)
      --username=''
            chart repository username where to locate the requested chart
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

