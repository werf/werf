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
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
            Command needs granted permissions to read, pull and delete images from the specified 
            stages storage and images repo
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images-repo='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; use WERF_STAGES_STORAGE environment by default).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

