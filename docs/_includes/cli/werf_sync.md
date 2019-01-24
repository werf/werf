{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Remove local stages cache for the images, that doesn't exist in the Docker registry.

Sync is a werf ability to automate periodical cleaning of build machine. Command should run after 
cleaning up Docker registry with the cleanup command.
See more info about sync: 
https://flant.github.io/werf/reference/registry/cleaning.html#local-storage-synchronization

Command should run from the project directory, where werf.yaml file reside.

{{ header }} Syntax

```bash
werf sync [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for sync
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --registry-password='':
            Docker registry password (granted read permission)
      --registry-username='':
            Docker registry username (granted read permission)
      --repo='':
            Docker repository name to get images information
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

{{ header }} Environments

```bash
  $WERF_DISABLE_SYNC_LOCAL_STAGES_DATE_PERIOD_POLICY  
  $WERF_HOME                                          
```

