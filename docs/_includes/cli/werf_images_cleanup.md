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
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
            Command needs granted permissions to delete images from the specified images repo.
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
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

