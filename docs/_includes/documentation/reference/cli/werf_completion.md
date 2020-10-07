{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate bash completion scripts

{{ header }} Syntax

```shell
werf completion [options]
```

{{ header }} Examples

```shell
  # Load bash completion
  $ source <(werf completion)

  # Load zsh completion
  $ autoload -Uz compinit && compinit -C
  $ source <(werf completion --shell=zsh)
```

{{ header }} Options

```shell
      --shell='bash'
            Set to bash or zsh (default $WERF_SHELL or bash)
```

