{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate werf environment variables for specified CI system.

Currently supported only GitLab CI

{{ header }} Syntax

```bash
werf ci-env CI_SYSTEM [options]
```

{{ header }} Examples

```bash
  # Load generated werf environment variables on gitlab job runner
  $ source <(werf ci-env gitlab --tagging-strategy tag-or-branch)
```

{{ header }} Options

```bash
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command will copy specified or default (~/.docker) config to the new temporary config 
            and may perform additional logins into new config.
  -h, --help=false:
            help for ci-env
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-repo=false:
            Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)
      --tagging-strategy='':
            tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified 
            CI_SYSTEM environment variables
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --verbose=false:
            Generate echo command for each resulted script line
```

