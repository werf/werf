{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Decrypt data from FILE_PATH or pipe.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file

{{ header }} Syntax

```shell
werf helm secret file decrypt [FILE_PATH] [options]
```

{{ header }} Examples

```shell
  # Decrypt secret file
  $ werf helm secret file decrypt .helm/secret/privacy

  # Decrypt from a pipe
  $ cat .helm/secret/date | werf helm secret decrypt
  Tue Jun 26 09:58:10 PDT 1990
```

{{ header }} Environments

```shell
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy. Recommended way to  
                    set secret key in CI-system. 
                    
                    Secret key also can be defined in files:
                    * ~/.werf/global_secret_key (globally),
                    * .werf_secret_key (per project)
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for decrypt
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
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
  -o, --output-file-path='':
            Write to file instead of stdout
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

