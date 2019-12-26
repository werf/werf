{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Regenerate secret files with new secret key.

Old key should be specified in the $WERF_OLD_SECRET_KEY.
New key should reside either in the $WERF_SECRET_KEY or .werf_secret_key file.

Command will extract data with the old key, generate new secret data and rewrite files:
* standard raw secret files in the .helm/secret folder;
* standard secret values yaml file .helm/secret-values.yaml;
* additional secret values yaml files specified with EXTRA_SECRET_VALUES_FILE_PATH params

{{ header }} Syntax

```shell
werf helm secret rotate-secret-key [EXTRA_SECRET_VALUES_FILE_PATH...] [options]
```

{{ header }} Environments

```shell
  $WERF_SECRET_KEY      Use specified secret key to extract secrets for the deploy. Recommended way 
                        to set secret key in CI-system. 
                        
                        Secret key also can be defined in files:
                        * ~/.werf/global_secret_key (globally),
                        * .werf_secret_key (per project)
  $WERF_OLD_SECRET_KEY  Use specified old secret key to rotate secrets
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for rotate-secret-key
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

