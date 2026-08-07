{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Retrieve a package from a package repository, and download it locally.

This is useful for fetching packages to inspect, modify, or repackage. It can
also be used to perform cryptographic verification of a chart without installing
the chart.

There are options for unpacking the chart after download. This will create a
directory for the chart and uncompress into that directory.

If the --verify flag is specified, the requested chart MUST have a provenance
file, and MUST pass the verification process. Failure in any part of this will
result in an error, and the chart will not be saved locally.


{{ header }} Syntax

```shell
werf helm pull [chart URL | repo/chartname] [...] [flags] [options]
```

{{ header }} Options

```shell
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify HTTPS client using this SSL certificate file
  -d, --destination="."
            location to write the chart. If this and untardir are specified, untardir is appended   
            to this
      --devel=false
            use development versions, too. Equivalent to version `>0.0.0-0`. If --version is set,   
            this is ignored.
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
      --prov=false
            fetch the provenance file, but don`t perform verification
      --repo=""
            chart repository url where to locate the requested chart
      --untar=false
            if set to true, will untar the chart after downloading it
      --untardir="."
            if untar is specified, this flag specifies the name of the directory into which the     
            chart is expanded
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

