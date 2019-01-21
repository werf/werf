Edit or create new secret file

```
werf secret edit FILE_PATH [options]
```

### Environments

```
  $WERF_SECRET_KEY  
  $WERF_TMP         
```

### Options

```
      --dir='': Change to the specified directory to find werf.yaml config
  -h, --help=false: help for edit
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=false: Edit FILE_PATH as secret values file
```

