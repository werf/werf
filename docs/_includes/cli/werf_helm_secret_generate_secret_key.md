{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate hex encryption key.
For further usage, the encryption key should be saved in $WERF_SECRET_KEY or .werf_secret_key file

{{ header }} Syntax

```shell
werf helm secret generate-secret-key [options]
```

{{ header }} Examples

```shell
  # Export encryption key
  $ export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

  # Save encryption key in .werf_secret_key file
  $ werf helm secret generate-secret-key > .werf_secret_key
```

{{ header }} Options

```shell
  -h, --help=false:
            help for generate-secret-key
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
```

