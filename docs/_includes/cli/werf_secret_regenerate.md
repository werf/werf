Regenerate secret files with new secret key

```
werf secret regenerate [EXTRA_SECRET_VALUES_FILE_PATH...] [options]
```

### Environments

```
  $WERF_SECRET_KEY  
```

### Options

```
      --dir='': Change to the specified directory to find werf.yaml config
  -h, --help=false: help for regenerate
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --old-key='': Old secret key
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

