{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
To load completion run

. <(werf completion)

To configure your bash shell to load completions for each session add to your ~/.bashrc ~/.profile

. <(werf completion)


{{ header }} Syntax

```bash
werf completion [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for completion
```

