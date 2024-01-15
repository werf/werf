{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Output shell completion code for the specified shell (Bash, Zsh, Fish, or PowerShell).
The shell code must be evaluated to provide interactive completion of `kubectl` commands. This can be done by sourcing it from the `.bash_profile`.

Detailed instructions on how to do this are available here:
* for macOS: [https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/#enable-shell-autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/#enable-shell-autocompletion);
* for linux: [https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion);

* for windows: [https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/#enable-shell-autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/#enable-shell-autocompletion);

> **Note for Zsh users**: Zsh completions are only supported in versions of Zsh >= 5.2.

{{ header }} Syntax

```shell
werf kubectl completion SHELL
```

{{ header }} Examples

```shell
  # Installing bash completion on macOS using homebrew
  ## If running Bash 3.2 included with macOS
  brew install bash-completion
  ## or, if running Bash 4.1+
  brew install bash-completion@2
  ## If kubectl is installed via homebrew, this should start working immediately
  ## If you've installed via other means, you may need add the completion to your completion directory
  kubectl completion bash > $(brew --prefix)/etc/bash_completion.d/kubectl
  
  
  # Installing bash completion on Linux
  ## If bash-completion is not installed on Linux, install the 'bash-completion' package
  ## via your distribution's package manager.
  ## Load the kubectl completion code for bash into the current shell
  source <(kubectl completion bash)
  ## Write bash completion code to a file and source it from .bash_profile
  kubectl completion bash > ~/.kube/completion.bash.inc
  printf "
  # kubectl shell completion
  source '$HOME/.kube/completion.bash.inc'
  " >> $HOME/.bash_profile
  source $HOME/.bash_profile
  
  # Load the kubectl completion code for zsh[1] into the current shell
  source <(kubectl completion zsh)
  # Set the kubectl completion code for zsh[1] to autoload on startup
  kubectl completion zsh > "${fpath[1]}/_kubectl"
  
  
  # Load the kubectl completion code for fish[2] into the current shell
  kubectl completion fish | source
  # To load completions for each session, execute once:
  kubectl completion fish > ~/.config/fish/completions/kubectl.fish
  
  # Load the kubectl completion code for powershell into the current shell
  kubectl completion powershell | Out-String | Invoke-Expression
  # Set kubectl completion code for powershell to run on startup
  ## Save completion code to a script and execute in the profile
  kubectl completion powershell > $HOME\.kube\completion.ps1
  Add-Content $PROFILE "$HOME\.kube\completion.ps1"
  ## Execute completion code in the profile
  Add-Content $PROFILE "if (Get-Command kubectl -ErrorAction SilentlyContinue) {
  kubectl completion powershell | Out-String | Invoke-Expression
  }"
  ## Add completion code directly to the $PROFILE script
  kubectl completion powershell >> $PROFILE
```

{{ header }} Options inherited from parent commands

```shell
      --as=''
            Username to impersonate for the operation. User could be a regular user or a service    
            account in a namespace.
      --as-group=[]
            Group to impersonate for the operation, this flag can be repeated to specify multiple   
            groups.
      --as-uid=''
            UID to impersonate for the operation.
      --cache-dir='~/.kube/cache'
            Default cache directory
      --certificate-authority=''
            Path to a cert file for the certificate authority
      --client-certificate=''
            Path to a client certificate file for TLS
      --client-key=''
            Path to a client key file for TLS
      --cluster=''
            The name of the kubeconfig cluster to use
      --context=''
            The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)
      --disable-compression=false
            If true, opt-out of response compression for all requests to the server
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kubeconfig=''
            Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or      
            $WERF_KUBECONFIG, or $KUBECONFIG). Ignored if kubeconfig passed as base64.
      --match-server-version=false
            Require server version to match client version
  -n, --namespace=''
            If present, the namespace scope for this CLI request
      --password=''
            Password for basic authentication to the API server
      --profile='none'
            Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)
      --profile-output='profile.pprof'
            Name of the file to write the profile to
      --request-timeout='0'
            The length of time to wait before giving up on a single server request. Non-zero values 
            should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don`t 
            timeout requests.
  -s, --server=''
            The address and port of the Kubernetes API server
      --tls-server-name=''
            Server name to use for server certificate validation. If it is not provided, the        
            hostname used to contact the server is used
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --token=''
            Bearer token for authentication to the API server
      --user=''
            The name of the kubeconfig user to use
      --username=''
            Username for basic authentication to the API server
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

