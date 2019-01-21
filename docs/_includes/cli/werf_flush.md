Delete project images in local docker storage and specified docker registry

```
werf flush [options]
```

### Environments

```
  $WERF_INSECURE_REGISTRY  
  $WERF_HOME               
```

### Options

```
      --dir='': Change to the specified directory to find werf.yaml config
      --dry-run=false: Indicate what the command would do without actually doing that
  -h, --help=false: help for flush
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --registry-password='': Docker registry password (granted read-write permission)
      --registry-username='': Docker registry username (granted read-write permission)
      --repo='': Docker repository name
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --with-dimgs=false: Delete images (not only stages cache)
```

