{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
add a chart repository

{{ header }} Syntax

```shell
werf helm repo add [NAME] [URL] [flags] [options]
```

{{ header }} Options

```shell
      --allow-deprecated-repos=false
            by default, this command will not allow adding official repos that have been            
            permanently deleted. This disables that behavior
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify HTTPS client using this SSL certificate file
      --force-update=false
            replace (overwrite) the repo if it already exists
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the repository
      --key-file=""
            identify HTTPS client using this SSL key file
      --pass-credentials=false
            pass credentials to all domains
      --password=""
            chart repository password
      --password-stdin=false
            read chart repository password from stdin
      --timeout=2m0s
            time to wait for the index file download to complete
      --username=""
            chart repository username
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

