{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup unused images from project images repo and stages storage.

This is the main cleanup command for periodical automated images cleaning. Command is supposed to 
be called daily for the project.

First step is 'werf images cleanup' command, which will delete unused images from images repo. 
Second step is 'werf stages cleanup' command, which will delete unused stages from stages repo (or 
locally) to be in sync with the images repo

{{ header }} Syntax

```bash
werf cleanup [options]
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
  $WERF_DOCKER_CONFIG                              Force usage of the specified docker config
  $WERF_INSECURE_REPO                              Enable insecure docker repo
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
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
  -p, --images-password='':
            Docker repo password (granted permission to read images info and delete images).
            Use by default:
            * WERF_CLEANUP_IMAGES_PASSWORD or 
            * WERF_IMAGES_PASSWORD environment
  -u, --images-username='':
            Images Docker repo username (granted permission to read images info and delete images).
            Use by default:
            * werf-cleanup username if WERF_IMAGES_PASSWORD environment defined or 
            * WERF_IMAGES_USERNAME environment
  -s, --stages='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now)
      --stages-password='':
            Stages Docker repo password
      --stages-username='':
            Stages Docker repo username
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --without-kube=false:
            Do not skip deployed kubernetes images
```

