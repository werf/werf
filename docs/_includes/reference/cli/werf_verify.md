{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Verify the integrity of built images by checking signatures of manifests and ELF files using specific image references.

The command outputs detailed logs about the verification process.

{{ header }} Syntax

```shell
werf verify --image-ref <image-ref> [--image-ref <image-ref>...] --verify-roots=<root-ca.pem> [--verify-manifest] [--verify-elf-files] [options]
```

{{ header }} Examples

```shell
  # Verify image manifest and ELF files for specific reference from Docker Hub
  $ werf verify --image-ref <DOCKER HUB USERNAME>/werf-guide-app:f4caaa836701e5346c4a0514bb977362ba5fe4ae114d0176f6a6c8cc-1612277803607 --verify-roots=/tmp/root-ca.pem --verify-manifest --verify-elf-files
```

{{ header }} Environments

```shell
  $WERF_DEBUG_ANSIBLE_ARGS  Pass specified cli args to ansible ($ANSIBLE_ARGS)
```

{{ header }} Options

```shell
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --image-ref=[]
            Verify only passed references (default $WERF_IMAGE_REF separated by comma)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --log-color-mode="auto"
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-project-dir=false
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-quiet=false
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format="2006-01-02T15:04:05Z07:00"
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
  -p, --parallel=true
            Run in parallel (default $WERF_PARALLEL or true)
      --parallel-tasks-limit=5
            Parallel tasks limit, set -1 to remove the limitation (default                          
            $WERF_PARALLEL_TASKS_LIMIT or 5)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --verify-bsign-elf-files=false
            Enable ELF files verification for bsign signed files (default                           
            $WERF_VERIFY_BSIGN_ELF_FILES).
      --verify-elf-files=false
            Enable ELF files verification (default $WERF_VERIFY_ELF_FILES).
            When enabled, the root certificates must be specified with --verify-roots option
      --verify-manifest=false
            Enable image manifest verification (default $WERF_VERIFY_MANIFEST).
            When enabled,
            the root certificates must be specified with --verify-roots option
      --verify-roots=[]
            The root certificates as path to PEM file or base64-encoded PEM (default                
            $WERF_VERIFY_ROOTS separated by comma)
```

