{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup is a werf ability to automate periodical cleaning of a docker registry.

Command deletes unused and old images from Docker registry by policies.
See more info about cleanup: https://flant.github.io/werf/reference/registry/cleaning.html#cleanup

Command should run from the project directory, where werf.yaml file reside.

{{ header }} Syntax

```bash
werf cleanup [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --registry-password='':
            Docker registry password (granted read-write permission)
      --registry-username='':
            Docker registry username (granted read-write permission)
      --repo='':
            Docker repository name
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

{{ header }} Environments

```bash
  $WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY     
  $WERF_GIT_TAGS_LIMIT_POLICY                  
  $WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY  
  $WERF_GIT_COMMITS_LIMIT_POLICY               
  $WERF_CLEANUP_REGISTRY_PASSWORD              
  $WERF_DOCKER_CONFIG                          
  $WERF_IGNORE_CI_DOCKER_AUTOLOGIN             
  $WERF_INSECURE_REGISTRY                      
  $WERF_HOME                                   
```

