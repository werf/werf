{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate Werf chart which will contain a valid Helm chart to the specified path.

Werf will generate additional values files, templates Chart.yaml and other files specific to the 
Werf chart. The result is a valid Helm chart

{{ header }} Syntax

```bash
werf helm generate-chart PATH [options]
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY     Use specified secret key to extract secrets for the deploy; recommended way 
                       to set secret key in CI-system
  $WERF_DOCKER_CONFIG  Force usage of the specified docker config
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (use WERF_DEPLOY_ENVIRONMENT by default)
  -h, --help=false:
            help for generate-chart
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
  -p, --images-password='':
            Images Docker repo username (granted permission to read images info, use 
            WERF_IMAGES_PASSWORD environment by default)
  -u, --images-username='':
            Images Docker repo username (granted permission to read images info, use 
            WERF_IMAGES_USERNAME environment by default)
      --namespace='':
            Use specified Kubernetes namespace (use %project-%environment template by default)
      --secret-values=[]:
            Additional helm secret values
      --ssh-key=[]:
            Enable only specified ssh keys (use system ssh-agent by default)
  -s, --stages='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now)
      --stages-password='':
            Stages Docker repo password
      --stages-username='':
            Stages Docker repo username
      --tag=[]:
            Add tag (can be used one or more times)
      --tag-git-branch='':
            Tag by git branch (use WERF_AUTOTAG_GIT_BRANCH environment by default)
      --tag-git-commit='':
            Tag by git commit (use WERF_AUTOTAG_GIT_COMMIT environment by default)
      --tag-git-tag='':
            Tag by git tag (use WERF_AUTOTAG_GIT_TAG environment by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

