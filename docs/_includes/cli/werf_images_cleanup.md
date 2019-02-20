{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup project images from images repo

{{ header }} Syntax

```bash
werf images cleanup [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to delete images from the specified images repo.
      --dry-run=false:
            Indicate what the command would do without actually doing that
      --git-commit-strategy-expiry-days=-1:
            Keep images published with the git-commit tagging strategy in the images repo for the 
            specified maximum days since image published. Republished image will be kept specified 
            maximum days since new publication date. No days limit by default, -1 disables the 
            limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS environment 
            variable.
      --git-commit-strategy-limit=-1:
            Keep max number of images published with the git-commit tagging strategy in the images 
            repo. No limit by default, -1 disables the limit. Value can be specified by the 
            $WERF_GIT_COMMIT_STRATEGY_LIMIT environment variable.
      --git-tag-strategy-expiry-days=-1:
            Keep images published with the git-tag tagging strategy in the images repo for the 
            specified maximum days since image published. Republished image will be kept specified 
            maximum days since new publication date. No days limit by default, -1 disables the 
            limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS environment 
            variable.
      --git-tag-strategy-limit=-1:
            Keep max number of images published with the git-tag tagging strategy in the images 
            repo. No limit by default, -1 disables the limit. Value can be specified by the 
            $WERF_GIT_TAG_STRATEGY_LIMIT environment variable.
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME environment 
            or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO environment)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP environment or system 
            tmp dir)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

