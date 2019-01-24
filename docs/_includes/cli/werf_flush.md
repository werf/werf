{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete project images in local Docker storage and specified Docker registry.

This command is useful to fully delete all data related to the project from:
* local Docker storage or
* both local Docker storage and Docker registry if --repo parameter has been specified.
See more info about flush: https://flant.github.io/werf/reference/registry/cleaning.html#flush.

Command should run from the project directory, where werf.yaml file reside.

Flush requires read-write permissions to delete images from Docker registry. Standard Docker config 
or specified options --registry-username and --registry-password will be used to authorize in the 
Docker registry.

See more info about authorization: 
https://flant.github.io/werf/reference/registry/authorization.html

{{ header }} Syntax

```bash
werf flush [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for flush
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --registry-password='':
            Docker registry password (granted read-write permission)
      --registry-username='':
            Docker registry username (granted read-write permission)
      --repo='':
            Docker repository name
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --with-images=false:
            Delete images (not only stages cache)
```

{{ header }} Environments

```bash
  $WERF_INSECURE_REGISTRY  
  $WERF_HOME               
```

