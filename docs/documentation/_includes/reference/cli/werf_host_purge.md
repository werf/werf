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
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-mode='simple'
            Set development mode (default $WERF_DEV_MODE or simple).
            Two development modes are supported:
            - simple: for working with the worktree state of the git repository
            - strict: for working with the index state of the git repository
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
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info https://werf.io/documentation/advanced/giterminism.html, default              
            $WERF_LOOSE_GITERMINISM)
  -N, --project-name=''
            Set a specific project name (default $WERF_PROJECT_NAME)
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single repo.
            
            Default:
             - $WERF_SYNCHRONIZATION, or
             - :local if --repo is not specified, or
             - https://synchronization.werf.io if --repo has been specified.
            
            The same address should be specified for all werf processes that work with a single     
            repo. :local address allows execution of werf processes from a single host only
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

