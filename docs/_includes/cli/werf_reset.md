{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete all images, containers, and cache files for all projects created by werf on the host.

Reset is the fullest method of cleaning on the local machine.

No project files (i.e. werf.yaml) are needed to run reset.

See more info about reset type of cleaning: 
https://flant.github.io/werf/reference/registry/cleaning.html#reset

{{ header }} Syntax

```bash
werf reset [options]
```

{{ header }} Options

```bash
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for reset
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --only-cache-version=false:
            Only delete stages cache, images, and containers created by these werf versions which 
            are incompatible with current werf version
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

