{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Manage the dependencies of a chart.

Helm charts store their dependencies in &#39;charts/&#39;. For chart developers, it is
often easier to manage dependencies in &#39;Chart.yaml&#39; which declares all
dependencies.

The dependency commands operate on that file, making it easy to synchronize
between the desired dependencies and the actual dependencies stored in the
&#39;charts/&#39; directory.

For example, this Chart.yaml declares two dependencies:

    # Chart.yaml
    dependencies:
    - name: nginx
      version: &#34;1.2.3&#34;
      repository: &#34;[https://example.com/charts&#34](https://example.com/charts&#34);
    - name: memcached
      version: &#34;3.2.1&#34;
      repository: &#34;[https://another.example.com/charts&#34](https://another.example.com/charts&#34);


The &#39;name&#39; should be the name of a chart, where that name must match the name
in that chart&#39;s &#39;Chart.yaml&#39; file.

The &#39;version&#39; field should contain a semantic version or version range.

The &#39;repository&#39; URL should point to a Chart Repository. Helm expects that by
appending &#39;/index.yaml&#39; to the URL, it should be able to retrieve the chart
repository&#39;s index. Note: &#39;repository&#39; can be an alias. The alias must start
with &#39;alias:&#39; or &#39;@&#39;.

Starting from 2.2.0, repository can be defined as the path to the directory of
the dependency charts stored locally. The path should start with a prefix of
&#34;[file://&#34](file://&#34);. For example,

    # Chart.yaml
    dependencies:
    - name: nginx
      version: &#34;1.2.3&#34;
      repository: &#34;[file://../dependency_chart/nginx&#34](file://../dependency_chart/nginx&#34);

If the dependency chart is retrieved locally, it is not required to have the
repository added to helm by &#34;helm add repo&#34;. Version matching is also supported
for this case.


{{ header }} Options inherited from parent commands

```shell
      --hooks-status-progress-period=5
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or           
            $KUBECONFIG)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=''
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
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
  -n, --namespace=''
            namespace scope for this request
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

