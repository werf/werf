Remove local stages cache for the images, that don't exist into the docker registry

```
werf sync [options]
```

### Environments

```
  $WERF_DISABLE_SYNC_LOCAL_STAGES_DATE_PERIOD_POLICY  
  $WERF_HOME                                          
```

### Options

```
      --dir='': Change to the specified directory to find werf.yaml config
      --dry-run=false: Indicate what the command would do without actually doing that
  -h, --help=false: help for sync
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --registry-password='': Docker registry password (granted read permission)
      --registry-username='': Docker registry username (granted read permission)
      --repo='': Docker repository name to get images information
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

