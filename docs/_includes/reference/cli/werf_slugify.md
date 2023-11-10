{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print slugged string by specified format.

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
  -f, --format=''
              r|helm-release:         suitable for Helm Release
             ns|kubernetes-namespace: suitable for Kubernetes Namespace
            tag|docker-tag:           suitable for Docker Tag
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
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format='2006-01-02T15:04:05Z07:00'
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
```

