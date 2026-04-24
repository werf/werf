{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Authenticate to a remote registry.

For example for Github Container Registry:

    echo "$GITHUB_TOKEN" | helm registry login [ghcr.io](ghcr.io) -u $GITHUB_USER --password-stdin


{{ header }} Syntax

```shell
werf helm registry login [host] [flags] [options]
```

{{ header }} Options

```shell
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify registry client using this SSL certificate file
      --insecure=false
            allow connections to TLS registry without certs
      --key-file=""
            identify registry client using this SSL key file
  -p, --password=""
            registry password or identity token
      --password-stdin=false
            read password or identity token from stdin
      --plain-http=false
            use insecure HTTP connections for the chart upload
  -u, --username=""
            registry username
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

