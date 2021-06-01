{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command upgrades a release to a new version of a chart.

The upgrade arguments must be a release and chart. The chart
argument can be either: a chart reference(&#39;example/mariadb&#39;), a path to a chart directory,
a packaged chart, or a fully qualified URL. For chart references, the latest
version will be specified unless the &#39;--version&#39; flag is set.

To override values in a chart, use either the &#39;--values&#39; flag and pass in a file
or use the &#39;--set&#39; flag and pass configuration from the command line, to force string
values, use &#39;--set-string&#39;. In case a value is large and therefore
you want not to use neither &#39;--values&#39; nor &#39;--set&#39;, use &#39;--set-file&#39; to read the
single large value from file.

You can specify the &#39;--values&#39;/&#39;-f&#39; flag multiple times. The priority will be given to the
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml
contained a key called &#39;Test&#39;, the value set in override.yaml would take precedence:

    $ helm upgrade -f myvalues.yaml -f override.yaml redis ./redis

You can specify the &#39;--set&#39; flag multiple times. The priority will be given to the
last (right-most) set specified. For example, if both &#39;bar&#39; and &#39;newbar&#39; values are
set for a key called &#39;foo&#39;, the &#39;newbar&#39; value would take precedence:

    $ helm upgrade --set foo=bar --set foo=newbar redis ./redis


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
      --description=''
            add a custom description
      --devel=false
            use development versions, too. Equivalent to version `>0.0.0-0`. If --version is set,   
            this is ignored
      --disable-openapi-validation=false
            if set, the upgrade process will not validate rendered templates against the Kubernetes 
            OpenAPI Schema
      --dry-run=false
            simulate an upgrade
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
      --no-hooks=false
            disable pre/post upgrade hooks
  -o, --output=table
            prints the output in the specified format. Allowed values: table, json, yaml
      --password=''
            chart repository password where to locate the requested chart
      --post-renderer=exec
            the path to an executable to be used for post rendering. If it exists in $PATH, the     
            binary will be used, otherwise it will try to look for the executable at the given path
      --render-subchart-notes=false
            if set, render subchart notes along with the parent
      --repo=''
            chart repository url where to locate the requested chart
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
            Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or           
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
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
  -n, --namespace=''
            namespace scope for this request
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

