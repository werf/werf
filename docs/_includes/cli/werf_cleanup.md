{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Safely cleanup unused project images and stages.

First step is 'werf images cleanup' command, which will delete unused images from images repo. 
Second step is 'werf stages cleanup' command, which will delete unused stages from stages storage 
to be in sync with the images repo.

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel 
with other werf commands such as build, deploy and host cleanup.

{{ header }} Syntax

```bash
werf cleanup [options]
```

{{ header }} Examples

```bash
  $ werf cleanup --stages-storage :local --images-repo registry.mydomain.com/myproject
```

{{ header }} Environments

```bash
  $WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY  Redefine default 
  $WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY         Redefine default tags expiry date period 
                                                   policy: keep images built for git tags, that 
                                                   are no older than 30 days since build time
  $WERF_GIT_TAGS_LIMIT_POLICY                      Redefine default tags limit policy: keep no 
                                                   more than 10 images built for git tags
  $WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY      Redefine default commits expiry date period 
                                                   policy: keep images built for git commits, that 
                                                   are no older than 30 days since build time
  $WERF_GIT_COMMITS_LIMIT_POLICY                   Redefine default commits limit policy: keep no 
                                                   more than 50 images built for git commits
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default WERF_DOCKER_CONFIG or DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to read, pull and delete images from the specified 
            stages storage and images repo
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
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; default WERF_STAGES_STORAGE environment).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

