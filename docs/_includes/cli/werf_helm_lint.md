{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run lint procedure for the werf chart

{{ header }} Syntax

```shell
werf helm lint [options]
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
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --env='':
            Use specified environment (default $WERF_ENV)
  -h, --help=false:
            help for lint
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false:
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --secret-values=[]:
            Specify helm secret values in a YAML file (can specify multiple)
      --set=[]:
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2)
      --set-string=[]:
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]:
            Specify helm values in a YAML file or a URL (can specify multiple)
```

