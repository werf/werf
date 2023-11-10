{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
This command packages a chart into a versioned chart archive file. If a path is given, this will look at that path for a chart (which must contain a `Chart.yaml` file) and then package that directory.

Versioned chart archives are used by Helm package repositories.

To sign a chart, use the `--sign` flag. In most cases, you should also provide `--keyring path/to/secret/keys` and `--key keyname`.
```
$ helm package --sign ./mychart --key mykey --keyring ~/.gnupg/secring.gpg
```
If `--keyring` is not specified, Helm usually defaults to the public keyring unless your environment is otherwise configured.

{{ header }} Syntax

```shell
werf helm package [CHART_PATH] [...] [flags] [options]
```

{{ header }} Options

```shell
      --app-version=''
            set the appVersion on the chart to this version
  -u, --dependency-update=false
            update dependencies from "Chart.yaml" to dir "charts/" before packaging
  -d, --destination='.'
            location to write the chart.
      --key=''
            name of the key to use when signing. Used if --sign is true
      --keyring='~/.gnupg/pubring.gpg'
            location of a public keyring
      --passphrase-file=''
            location of a file which contains the passphrase for the signing key. Use "-" in order  
            to read from stdin.
      --sign=false
            use a PGP private key to sign this package
      --version=''
            set the version on the chart to this semver version
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

