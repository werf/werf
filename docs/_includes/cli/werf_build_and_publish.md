{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build stages and final images using each specified tag with the tagging strategy and push into 
images repo.

Command combines 'werf stages build' and 'werf images publish'.

After stages has been built, new docker layer with service info about tagging strategy will be 
built for each tag of each image from werf.yaml. Images will be pushed into docker repo with the 
names IMAGES_REPO/IMAGE_NAME:TAG. See more info about images naming: 
https://werf.io/reference/registry/image_naming.html.

The result of build-and-publish command is stages in stages storage and named images pushed into 
the docker repo.

If one or more IMAGE_NAME parameters specified, werf will build images stages and publish only 
these images from werf.yaml

{{ header }} Syntax

```bash
werf build-and-publish [IMAGE_NAME...] [options]
```

{{ header }} Examples

```bash
  # Build and publish all images from werf.yaml into specified docker repo, built stages will be placed locally; tag images with the mytag tag using custom tagging strategy
  $ werf build-and-publish --stages-storage :local --images-repo registry.mydomain.com/myproject --tag-custom mytag

  # Build and publish all images from werf.yaml into minikube registry; tag images with the mybranch tag, using git-branch tagging strategy
  $ werf build-and-publish --stages-storage :local --images-repo :minikube --tag-git-branch mybranch

  # Build and publish with enabled drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf --stages-storage :local build-and-publish --introspect-error --images-repo :minikube --tag-git-branch mybranch

  # Set --stages-storage default value using $WERF_STAGES_STORAGE param and --images-repo default value using $WERF_IMAGE_REPO param
  $ export WERF_STAGES_STORAGE=:local
  $ export WERF_IMAGES_REPO=myregistry.mydomain.com/myproject
  $ werf build-and-publish --tag-git-tag v1.4.9
```

{{ header }} Environments

```bash
  $WERF_DEBUG_ANSIBLE_ARGS  Pass specified cli args to ansible ($ANSIBLE_ARGS)
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to read, pull and push images into the specified 
            stages storage, to push images into the specified images repo, to pull base images.
  -h, --help=false:
            help for build-and-publish
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --insecure-repo=false:
            Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)
      --introspect-before-error=false:
            Introspect failed stage in the clean state, before running all assembly instructions of 
            the stage
      --introspect-error=false:
            Introspect failed stage in the state, right after running failed assembly instruction
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a 
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or 
            true).
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --ssh-key=[]:
            Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see 
            https://werf.io/reference/toolbox/ssh.html). Option can be specified multiple times to 
            use multiple keys.
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/reference/build/stages_and_images.html
      --tag-custom=[]:
            Use custom tagging strategy and tag by the specified arbitrary tags. Option can be used 
            multiple times to produce multiple images with the specified tags.
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
```

