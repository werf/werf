Образы описываются с помощью директивы _image_: `image: <image name>`, с которой начинается описание образа в конфигурации.
Параметр _image name_ — строка с именем, по аналогии с именем образа в Docker:

```yaml
image: frontend
```

Если в файле конфигурации описывается только один образ, то он может быть безымянным:

```yaml
image: ~
```

Если в файле конфигурации описывается более одного образа, то **каждый образ** должен иметь собственное имя:

```yaml
image: frontend
...
---
image: backend
...
```

Образ может иметь несколько имен, указываемых в виде YAML-списка (это эквивалентно описанию нескольких одинаковых образов с разными именами):

```yaml
image: [main-front,main-back]
```

Имя образа может быть использовано в большинстве команд:
* [werf build \[IMAGE_NAME...\] \[options\]]({{ site.baseurl }}/documentation/reference/cli/werf_build.html)
* [werf run \[options\] \[IMAGE_NAME\] \[-- COMMAND ARG...\]]({{ site.baseurl }}/documentation/reference/cli/werf_run.html)
