{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Debug cluster resources using interactive debugging containers.

`debug` provides automation for common debugging tasks for cluster objects identified by resource and name. Pods will be used by default if no resource is specified.

The action taken by `debug` varies depending on what resource is specified. Supported actions include:
* Workload: Create a copy of an existing pod with certain attributes changed, for example changing the image tag to a new version.
* Workload: Add an ephemeral container to an already running pod, for example to add debugging utilities without restarting the pod.
* Node: Create a new pod that runs in the node's host namespaces and can access the node's filesystem.

{{ header }} Syntax

```shell
werf kubectl debug (POD | TYPE[[.VERSION].GROUP]/NAME) [ -- COMMAND [args...] ] [options]
```

{{ header }} Examples

```shell
  # Create an interactive debugging session in pod mypod and immediately attach to it.
  kubectl debug mypod -it --image=busybox
  
  # Create an interactive debugging session for the pod in the file pod.yaml and immediately attach to it.
  # (requires the EphemeralContainers feature to be enabled in the cluster)
  kubectl debug -f pod.yaml -it --image=busybox
  
  # Create a debug container named debugger using a custom automated debugging image.
  kubectl debug --image=myproj/debug-tools -c debugger mypod
  
  # Create a copy of mypod adding a debug container and attach to it
  kubectl debug mypod -it --image=busybox --copy-to=my-debugger
  
  # Create a copy of mypod changing the command of mycontainer
  kubectl debug mypod -it --copy-to=my-debugger --container=mycontainer -- sh
  
  # Create a copy of mypod changing all container images to busybox
  kubectl debug mypod --copy-to=my-debugger --set-image=*=busybox
  
  # Create a copy of mypod adding a debug container and changing container images
  kubectl debug mypod -it --copy-to=my-debugger --image=debian --set-image=app=app:debug,sidecar=sidecar:debug
  
  # Create an interactive debugging session on a node and immediately attach to it.
  # The container will run in the host namespaces and the host's filesystem will be mounted at /host
  kubectl debug node/mynode -it --image=busybox
```

{{ header }} Options

```shell
      --arguments-only=false
            If specified, everything after -- will be passed to the new container as Args instead   
            of Command.
      --attach=false
            If true, wait for the container to start running, and then attach as if `kubectl attach 
            ...` were called.  Default false, unless `-i/--stdin` is set, in which case the default 
            is true.
  -c, --container=''
            Container name to use for debug container.
      --copy-to=''
            Create a copy of the target Pod with this name.
      --env=[]
            Environment variables to set in the container.
  -f, --filename=[]
            identifying the resource to debug
      --image=''
            Container image to use for debug container.
      --image-pull-policy=''
            The image pull policy for the container. If left empty, this value will not be          
            specified by the client and defaulted by the server.
      --profile='legacy'
            Debugging profile. Options are "legacy", "general", "baseline", "netadmin", or          
            "restricted".
  -q, --quiet=false
            If true, suppress informational messages.
      --replace=false
            When used with `--copy-to`, delete the original Pod.
      --same-node=false
            When used with `--copy-to`, schedule the copy of target Pod on the same node.
      --set-image=[]
            When used with `--copy-to`, a list of name=image pairs for changing container images,   
            similar to how `kubectl set image` works.
      --share-processes=true
            When used with `--copy-to`, enable process namespace sharing in the copy.
  -i, --stdin=false
            Keep stdin open on the container(s) in the pod, even if nothing is attached.
      --target=''
            When using an ephemeral container, target processes in this container name.
  -t, --tty=false
            Allocate a TTY for the debugging container.
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

