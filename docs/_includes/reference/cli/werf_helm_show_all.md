{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command inspects a chart (directory, file, or URL) and displays all its content
(values.yaml, Chart.yaml, README)


{{ header }} Syntax

```shell
werf helm show all [CHART] [flags] [options]
```

{{ header }} Options

```shell
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify HTTPS client using this SSL certificate file
      --devel=false
            use development versions, too. Equivalent to version `>0.0.0-0`. If --version is set,   
            this is ignored
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the chart download
      --key-file=""
            identify HTTPS client using this SSL key file
      --keyring="~/.gnupg/pubring.gpg"
            location of public keys used for verification
      --pass-credentials=false
            pass credentials to all domains
      --password=""
            chart repository password where to locate the requested chart
      --plain-http=false
            use insecure HTTP connections for the chart download
      --repo=""
            chart repository url where to locate the requested chart
      --username=""
            chart repository username where to locate the requested chart
      --verify=false
            verify the package before using it
      --version=""
            specify a version constraint for the chart version to use. This constraint can be a     
            specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is   
            not specified, the latest version is used
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

