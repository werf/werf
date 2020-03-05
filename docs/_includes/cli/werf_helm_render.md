{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Render werf chart templates to stdout

{{ header }} Syntax

```shell
werf helm render [options]
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
            ~/.docker (in the order of priority)
      --env='':
            Use specified environment (default $WERF_ENV)
  -h, --help=false:
            help for render
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false:
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --images-repo-mode='multirepo':
            Define how to store images in Repo: multirepo or monorepo (defaults to                  
            $WERF_IMAGES_REPO_MODE or multirepo)
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
      --namespace='':
            Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or         
            deploy.namespace custom template from werf.yaml)
  -o, --output-file-path='':
            Write to file instead of stdout
      --release='':
            Use specified Helm release name (default [[ project ]]-[[ env ]] template or            
            deploy.helmRelease custom template from werf.yaml)
      --secret-values=[]:
            Specify helm secret values in a YAML file (can specify multiple)
      --set=[]:
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2)
      --set-string=[]:
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2)
      --tag-by-stages-signature=false:
            Use stages-signature tagging strategy and tag each image by the corresponding signature 
            of last image stage (option can be enabled by specifying                                
            $WERF_TAG_BY_STAGES_SIGNATURE=true)
      --tag-custom=[]:
            Use custom tagging strategy and tag by the specified arbitrary tags.
            Option can be used multiple times to produce multiple images with the specified tags.
            Also can be specified in $WERF_TAG_CUSTOM* (e.g. $WERF_TAG_CUSTOM_TAG1=tag1,            
            $WERF_TAG_CUSTOM_TAG2=tag2)
      --tag-git-branch='':
            Use git-branch tagging strategy and tag by the specified git branch (option can be      
            enabled by specifying git branch in the $WERF_TAG_GIT_BRANCH)
      --tag-git-commit='':
            Use git-commit tagging strategy and tag by the specified git commit hash (option can be 
            enabled by specifying git commit hash in the $WERF_TAG_GIT_COMMIT)
      --tag-git-tag='':
            Use git-tag tagging strategy and tag by the specified git tag (option can be enabled by 
            specifying git tag in the $WERF_TAG_GIT_TAG)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]:
            Specify helm values in a YAML file or a URL (can specify multiple)
```

