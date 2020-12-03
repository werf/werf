{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge werf images, cache and other data for all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, converge and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. This 
command is supposed to be run manually.

{{ header }} Syntax

```shell
werf host purge [options]
```

{{ header }} Options

```shell
      --disable-giterminism=false
            Disable werf giterminism mode (more info                                                
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_DISABLE_GITERMINISM)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --force=false
            First remove containers that use werf docker images which are going to be deleted
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --log-color-mode='auto'
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --non-strict-giterminism-inspection=false
            Change some errors to warnings during giterminism inspection (more info                 
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_NON_STRICT_GITERMINISM_INSPECTION)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

