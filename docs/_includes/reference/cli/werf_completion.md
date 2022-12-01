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
  # or for older bash versions (e.g. bash 3.2 on macOS):
  $ source /dev/stdin <<< "$(werf completion)"

  # Load zsh completion
  $ autoload -Uz compinit && compinit -C
  $ source <(werf completion --shell=zsh)

  # Load fish completion
  $ source <(werf completion --shell=fish)

  # Load powershell completion
  $ werf completion --shell=powershell | Out-String | Invoke-Expression
```

{{ header }} Options

```shell
      --shell='bash'
            Set to bash, zsh, fish or powershell (default $WERF_SHELL or bash)
```

