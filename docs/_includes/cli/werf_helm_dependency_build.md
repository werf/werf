{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Build out the charts/ directory from the requirements.lock file.

Build is used to reconstruct a chart's dependencies to the state specified in
the lock file.

If no lock file is found, 'werf helm dependency build' will mirror the behavior of
the 'werf helm dependency update' command. This means it will update the on-disk
dependencies to mirror the requirements.yaml file and generate a lock file.


{{ header }} Syntax

```shell
werf helm dependency build [options]
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for build
      --keyring='$HOME/.gnupg/pubring.gpg':
            keyring containing public keys
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false:
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false:
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false:
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --verify=false:
            verify the packages against signatures
```

