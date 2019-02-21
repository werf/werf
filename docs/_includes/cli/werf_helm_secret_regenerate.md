{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Regenerate secret files with new secret key.

Old key should be specified with the --old-key option.
New key should reside either in the $WERF_SECRET_KEY or .werf_secret_key file.

Command will extract data with the old key, generate new secret data and rewrite files:
* standard raw secret files in the .helm/secret folder;
* standard secret values yaml file .helm/secret-values.yaml;
* additional secret values yaml files specified with EXTRA_SECRET_VALUES_FILE_PATH params

{{ header }} Syntax

```bash
werf helm secret regenerate [EXTRA_SECRET_VALUES_FILE_PATH...] [options]
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy; recommended way to 
                    set secret key in CI-system
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --disable-pretty-log=false:
            Disable emojis, auto line wrapping and replace log process border characters with 
            spaces (default $WERF_DISABLE_PRETTY_LOG).
  -h, --help=false:
            help for regenerate
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a 
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --old-key='':
            Old secret key
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP or system tmp dir)
```

