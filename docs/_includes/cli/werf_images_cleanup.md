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

{{ header }} Environments

```bash
  $WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY     Redefine default tags expiry date period policy: 
                                               keep images built for git tags, that are no older 
                                               than 30 days since build time
  $WERF_GIT_TAGS_LIMIT_POLICY                  Redefine default tags limit policy: keep no more 
                                               than 10 images built for git tags
  $WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY  Redefine default commits expiry date period policy: 
                                               keep images built for git commits, that are no 
                                               older than 30 days since build time
  $WERF_GIT_COMMITS_LIMIT_POLICY               Redefine default commits limit policy: keep no more 
                                               than 50 images built for git commits
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default WERF_DOCKER_CONFIG or DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to delete images from the specified images repo.
      --dry-run=false:
            Indicate what the command would do without actually doing that
      --git-commit-strategy-expiry-days=0:
            Keep images published with the git-commit tagging strategy in the images repo for the 
            specified maximum days since image published. Republished image will be kept specified 
            maximum days since new publication date. No days limit by default. Value can be 
            specified by the WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS environment variable."
      --git-commit-strategy-limit=0:
            Keep max number of images published with the git-commit tagging strategy in the images 
            repo. No limit by default. Value can be specified by the 
            WERF_GIT_COMMIT_STRATEGY_LIMIT environment variable.
      --git-tag-strategy-expiry-days=0:
            Keep images published with the git-tag tagging strategy in the images repo for the 
            specified maximum days since image published. Republished image will be kept specified 
            maximum days since new publication date. No days limit by default. Value can be 
            specified by the WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS environment variable."
      --git-tag-strategy-limit=0:
            Keep max number of images published with the git-tag tagging strategy in the images 
            repo. No limit by default. Value can be specified by the WERF_GIT_TAG_STRATEGY_LIMIT 
            environment variable.
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default WERF_HOME environment or 
            ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default WERF_IMAGES_REPO environment)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

