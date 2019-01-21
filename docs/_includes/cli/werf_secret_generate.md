Generate secret data

```
werf secret generate [options]
```

### Environments

```
  $WERF_SECRET_KEY  
```

### Options

```
      --dir='': Change to the specified directory to find werf.yaml config
      --file-path='': Encode file data by specified path
  -h, --help=false: help for generate
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --output-file-path='': Save encoded data by specified file path
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=false: Encode specified FILE_PATH (--file-path) as secret values file
```

