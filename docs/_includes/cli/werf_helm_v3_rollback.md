{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command rolls back a release to a previous revision.

The first argument of the rollback command is the name of a release, and the
second is a revision (version) number. If this argument is omitted, it will
roll back to the previous release.

To see revision numbers, run 'helm history RELEASE'.


{{ header }} Syntax

```shell
werf helm-v3 rollback <RELEASE> [REVISION] [flags] [options]
```

{{ header }} Options

```shell
      --cleanup-on-fail=false:
            allow deletion of new resources created in this rollback when rollback fails
      --dry-run=false:
            simulate a rollback
      --force=false:
            force resource update through delete/recreate if needed
  -h, --help=false:
            help for rollback
      --hooks-status-progress-period=5:
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --kube-config='':
            Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or           
            $KUBECONFIG)
      --kube-config-base64='':
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
  -n, --namespace='':
            namespace scope for this request
      --no-hooks=false:
            prevent hooks from running during rollback
      --recreate-pods=false:
            performs pods restart for the resource if applicable
      --status-progress-period=5:
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
      --timeout=5m0s:
            time to wait for any individual Kubernetes operation (like Jobs for hooks)
      --wait=false:
            if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a       
            Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release  
            as successful. It will wait for as long as --timeout
```

