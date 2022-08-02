{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Help provides help for any command in the application.
Simply type werf help [path to command] for full details.

{{ header }} Syntax

```shell
werf help [command]
```

