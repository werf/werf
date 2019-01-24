{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Force run werf gc procedure.

Werf GC procedure typically runs automatically during execution of some command. With this command 
user can force werf to run gc procedure to force removal of excess tmp files right now, for example.

See more info about gc: https://flant.github.io/werf/reference/registry/cleaning.html#gc

{{ header }} Syntax

```bash
werf gc [options]
```

{{ header }} Options

```bash
  -h, --help=false:
            help for gc
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

