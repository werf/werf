{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Render Werf chart templates to stdout

{{ header }} Syntax

```bash
werf helm render [options]
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy. Recommended way to 
                    set secret key in CI-system. 
                    
                    Secret key also can be defined in files:
                    * ~/.werf/global_secret_key (globally),
                    * .werf_secret_key (per project)
```

{{ header }} Options

```bash
      --add-annotation=[]:
            Add annotation to deploying resources (can specify multiple).
            Format: annoName=annoValue.
            Also can be specified in $WERF_ADD_ANNOTATION* (e.g. 
            $WERF_ADD_ANNOTATION_1=annoName1=annoValue1", 
            $WERF_ADD_ANNOTATION_2=annoName2=annoValue2")
      --add-label=[]:
            Add label to deploying resources (can specify multiple).
            Format: labelName=labelValue.
            Also can be specified in $WERF_ADD_LABEL* (e.g. 
            $WERF_ADD_LABEL_1=labelName1=labelValue1", $WERF_ADD_LABEL_2=labelName2=labelValue2")
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
      --env='':
            Use specified environment (default $WERF_ENV)
  -h, --help=false:
            help for render
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

