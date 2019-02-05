{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge werf images, stages, cache and other data of all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, publish, deploy and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. 
This command is supposed to be run manually.

{{ header }} Syntax

```bash
werf host purge [options]
```

{{ header }} Options

```bash
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

