{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}


{{ header }} Syntax

```bash
werf deploy [options]
```

{{ header }} Options

```bash
      --dir='': Change to the specified directory to find werf.yaml config
      --environment='': Use specified environment (use CI_ENVIRONMENT_SLUG by default). Environment is a required parameter and should be specified with option or CI_ENVIRONMENT_SLUG variable.
  -h, --help=false: help for deploy
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --kube-context='': Kubernetes config context
      --namespace='': Use specified Kubernetes namespace (use %project-%environment template by default)
      --registry-password='': Docker registry password
      --registry-username='': Docker registry username
      --release='': Use specified Helm release name (use %project-%environment template by default)
      --repo='': Docker repository name to get images ids from. CI_REGISTRY_IMAGE will be used by default if available.
      --secret-values=[]: Additional helm secret values
      --set=[]: Additional helm sets
      --set-string=[]: Additional helm STRING sets
      --ssh-key=[]: Enable only specified ssh keys (use system ssh-agent by default)
      --tag=[]: Add tag (can be used one or more times)
      --tag-branch=false: Tag by git branch
      --tag-build-id=false: Tag by CI build id
      --tag-ci=false: Tag by CI branch and tag
      --tag-commit=false: Tag by git commit
  -t, --timeout=0: watch timeout in seconds
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=[]: Additional helm values
      --without-registry=false: Do not get images info from registry
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY                  
  $WERF_DOCKER_CONFIG               
  $WERF_IGNORE_CI_DOCKER_AUTOLOGIN  
  $WERF_HOME                        
  $WERF_TMP                         
```

