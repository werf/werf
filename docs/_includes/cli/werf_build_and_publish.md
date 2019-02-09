{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build stages then publish images into Docker repo.

New docker layer with service info about tagging scheme will be built for each image. Images will
be pushed into docker repo with the names IMAGE_REPO/IMAGE_NAME:TAG. See more info about images
naming: https://flant.github.io/werf/reference/registry/image_naming.html.

The result of build-and-publish command is a stages cache for images and named images pushed into
the docker repo.

If one or more IMAGE_NAME parameters specified, werf will build images stages and publish only
these images from werf.yaml

{{ header }} Syntax

```bash
werf build-and-publish [IMAGE_NAME...] [options]
```

{{ header }} Environments

```bash
  $WERF_DEBUG_ANSIBLE_ARGS   Pass specified cli args to ansible (ANSIBLE_ARGS)
  $WERF_DOCKER_CONFIG  Force usage of the specified docker config
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for build-and-publish
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
  -p, --images-password='':
            Images Docker repo password (granted permission to push images, use
            WERF_IMAGES_PASSWORD environment by default)
  -u, --images-username='':
            Images Docker repo username (granted permission to push images, use
            WERF_IMAGES_USERNAME environment by default)
      --introspect-before-error=false:
            Introspect failed stage in the clean state, before running all assembly instructions
            of the stage
      --introspect-error=false:
            Introspect failed stage in the state, right after running failed assembly instruction
      --pull-password='':
            Password to authorize pull of base images
      --pull-username='':
            Username to authorize pull of base images
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

