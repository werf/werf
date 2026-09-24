{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Update the on-disk dependencies to mirror Chart.yaml.

This command verifies that the required charts, as expressed in 'Chart.yaml',
are present in 'charts/' and are at an acceptable version. It will pull down
the latest charts that satisfy the dependencies, and clean up old dependencies.

On successful update, this will generate a lock file that can be used to
rebuild the dependencies to an exact version.

Dependencies are not required to be represented in 'Chart.yaml'. For that
reason, an update command will not remove charts unless they are (a) present
in the Chart.yaml file, but (b) at the wrong version.


{{ header }} Syntax

```shell
werf helm dependency update CHART [flags] [options]
```

{{ header }} Options

```shell
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify HTTPS client using this SSL certificate file
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the chart download
      --key-file=""
            identify HTTPS client using this SSL key file
      --keyring="~/.gnupg/pubring.gpg"
            keyring containing public keys
      --password=""
            chart repository password where to locate the requested chart
      --plain-http=false
            use insecure HTTP connections for the chart download
      --skip-refresh=false
            do not refresh the local repository cache
      --username=""
            chart repository username where to locate the requested chart
      --verify=false
            verify the packages against signatures
```

{{ header }} Options inherited from parent commands

```shell
      --log-color-mode="auto"
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
      --log-time-format="2006-01-02T15:04:05Z07:00"
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
```

