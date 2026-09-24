{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command packages a chart into a versioned chart archive file. If a path
is given, this will look at that path for a chart (which must contain a
Chart.yaml file) and then package that directory.

Versioned chart archives are used by Helm package repositories.

To sign a chart, use the '--sign' flag. In most cases, you should also
provide '--keyring path/to/secret/keys' and '--key keyname'.

  $ helm package --sign ./mychart --key mykey --keyring ~/.gnupg/secring.gpg

If '--keyring' is not specified, Helm usually defaults to the public keyring
unless your environment is otherwise configured.


{{ header }} Syntax

```shell
werf helm package [CHART_PATH] [...] [flags] [options]
```

{{ header }} Options

```shell
      --app-version=""
            set the appVersion on the chart to this version
      --ca-file=""
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file=""
            identify HTTPS client using this SSL certificate file
  -u, --dependency-update=false
            update dependencies from "Chart.yaml" to dir "charts/" before packaging
  -d, --destination="."
            location to write the chart.
      --insecure-skip-tls-verify=false
            skip tls certificate checks for the chart download
      --key=""
            name of the key to use when signing. Used if --sign is true
      --key-file=""
            identify HTTPS client using this SSL key file
      --keyring="~/.gnupg/pubring.gpg"
            location of a public keyring
      --passphrase-file=""
            location of a file which contains the passphrase for the signing key. Use "-" in order  
            to read from stdin.
      --password=""
            chart repository password where to locate the requested chart
      --plain-http=false
            use insecure HTTP connections for the chart download
      --sign=false
            use a PGP private key to sign this package
      --username=""
            chart repository username where to locate the requested chart
      --version=""
            set the version on the chart to this semver version
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

