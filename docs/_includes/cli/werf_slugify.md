{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print slugged string by specified format

{{ header }} Syntax

```shell
werf slugify STRING [options]
```

{{ header }} Examples

```shell
  $ werf slugify -f kubernetes-namespace feature-fix-2
  feature-fix-2

  $ werf slugify -f kubernetes-namespace 'branch/one/!@#4.4-3'
  branch-one-4-4-3-4fe08955

  $ werf slugify -f kubernetes-namespace My_branch
  my-branch-8ebf2d1d

  $ werf slugify -f helm-release my_release-NAME
  my_release-NAME

  # The result has been trimmed to fit maximum bytes limit:
  $ werf slugify -f helm-release looooooooooooooooooooooooooooooooooooooooooong_string
  looooooooooooooooooooooooooooooooooooooooooo-b150a895

  $ werf slugify -f docker-tag helo/ehlo
  helo-ehlo-b6f6ab1f

  $ werf slugify -f docker-tag 16.04
  16.04
```

{{ header }} Options

```shell
  -f, --format='':
              r|helm-release:         suitable for Helm Release
             ns|kubernetes-namespace: suitable for Kubernetes Namespace
            tag|docker-tag:           suitable for Docker Tag
  -h, --help=false:
            help for slugify
```

