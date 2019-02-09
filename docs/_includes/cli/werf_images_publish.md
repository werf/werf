{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build final images using each specified tag with the tagging strategy and push into images repo.

New docker layer with service info about tagging scheme will be built for each tag of each image 
from werf.yaml. Images will be pushed into docker repo with the names IMAGES_REPO/IMAGE_NAME:TAG. 
See more info about images naming: 
https://flant.github.io/werf/reference/registry/image_naming.html.

If one or more IMAGE_NAME parameters specified, werf will publish only these images from werf.yaml.

{{ header }} Syntax

```bash
werf images publish [IMAGE_NAME...] [options]
```

{{ header }} Environments

```bash
  $WERF_DOCKER_CONFIG  Force usage of the specified docker config
  $WERF_INSECURE_REPO  Enable insecure docker repo
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
  -h, --help=false:
            help for publish
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
      --ssh-key=[]:
            Use only specific ssh keys (system ssh-agent or default keys will be used by default, 
            see https://flant.github.io/werf/reference/toolbox/ssh.html). Option can be specified 
            multiple times to use multiple keys.
  -s, --stages='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now)
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

