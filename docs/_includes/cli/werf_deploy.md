{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Deploy application into Kubernetes.

Command will create Helm Release and wait until all resources of the release are become ready.

Deploy needs the same parameters as push to construct image names: repo and tags. Docker images 
names are constructed from paramters as IMAGES_REPO/IMAGE_NAME:TAG. Deploy will fetch built image 
ids from Docker repo. So images should be published prior running deploy.

Helm chart directory .helm should exists and contain valid Helm chart.

Environment is a required param for the deploy by default, because it is needed to construct Helm 
Release name and Kubernetes Namespace. Either --env or WERF_DEPLOY_ENVIRONMENT should be specified 
for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to 
change it: https://flant.github.io/werf/reference/deploy/deploy_to_kubernetes.html

{{ header }} Syntax

```bash
werf deploy [options]
```

{{ header }} Examples

```bash
  # Deploy project named 'myproject' into 'dev' environment using images from registry.mydomain.com/myproject tagged as mytag with git-tag tagging strategy; helm release name and namespace will be named as 'myproject-dev'
  $ werf deploy --env dev --stages-storage :local --images-repo registry.mydomain.com/myproject --tag-git-tag mytag

  # Deploy project using specified helm release name and namespace using images from registry.mydomain.com/myproject
  $ werf deploy --release myrelease --namespace myns --stages-storage :local --images-repo registry.mydomain.com/myproject
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy; recommended way to 
                    set secret key in CI-system
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
            Command needs granted permissions to read and pull images from the specified stages 
            storage and images repo
      --env='':
            Use specified environment (use WERF_DEPLOY_ENVIRONMENT by default)
  -h, --help=false:
            help for deploy
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images-repo='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context
      --namespace='':
            Use specified Kubernetes namespace (use %project-%environment template by default)
      --release='':
            Use specified Helm release name (use %project-%environment template by default)
      --secret-values=[]:
            Additional helm secret values
      --set=[]:
            Additional helm sets
      --set-string=[]:
            Additional helm STRING sets
      --ssh-key=[]:
            Use only specific ssh keys (system ssh-agent or default keys will be used by default, 
            see https://flant.github.io/werf/reference/toolbox/ssh.html). Option can be specified 
            multiple times to use multiple keys.
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; use WERF_STAGES_STORAGE environment by default).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tag-custom=[]:
            Use custom tagging strategy and tag by the specified arbitrary tags. Option can be 
            used multiple times to produce multiple images with the specified tags.
      --tag-git-branch='':
            Use git-branch tagging strategy and tag by the specified git branch (option can be 
            enabled by specifying git branch in the WERF_TAG_GIT_BRANCH environment variable)
      --tag-git-commit='':
            Use git-commit tagging strategy and tag by the specified git commit hash (option can 
            be enabled by specifying git commit hash in the WERF_TAG_GIT_COMMIT environment 
            variable)
      --tag-git-tag='':
            Use git-tag tagging strategy and tag by the specified git tag (option can be enabled 
            by specifying git tag in the WERF_TAG_GIT_TAG environment variable)
  -t, --timeout=0:
            Resources tracking timeout in seconds
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --values=[]:
            Additional helm values
```

