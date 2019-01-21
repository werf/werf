Delete images, containers, and cache files for all projects created by werf on the host

```
werf reset [options]
```

### Options

```
      --dry-run=false: Indicate what the command would do without actually doing that
  -h, --help=false: help for reset
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --only-cache-version=false: Only delete stages cache, images, and containers created by another werf version
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

