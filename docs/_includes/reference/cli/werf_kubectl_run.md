{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Create and run a particular image in a pod.

{{ header }} Syntax

```shell
werf kubectl run NAME --image=image [--env="key=value"] [--port=port] [--dry-run=server|client] [--overrides=inline-json] [--command] -- [COMMAND] [args...] [options]
```

{{ header }} Examples

```shell
  # Start a nginx pod
  kubectl run nginx --image=nginx
  
  # Start a hazelcast pod and let the container expose port 5701
  kubectl run hazelcast --image=hazelcast/hazelcast --port=5701
  
  # Start a hazelcast pod and set environment variables "DNS_DOMAIN=cluster" and "POD_NAMESPACE=default" in the container
  kubectl run hazelcast --image=hazelcast/hazelcast --env="DNS_DOMAIN=cluster" --env="POD_NAMESPACE=default"
  
  # Start a hazelcast pod and set labels "app=hazelcast" and "env=prod" in the container
  kubectl run hazelcast --image=hazelcast/hazelcast --labels="app=hazelcast,env=prod"
  
  # Dry run; print the corresponding API objects without creating them
  kubectl run nginx --image=nginx --dry-run=client
  
  # Start a nginx pod, but overload the spec with a partial set of values parsed from JSON
  kubectl run nginx --image=nginx --overrides='{ "apiVersion": "v1", "spec": { ... } }'
  
  # Start a busybox pod and keep it in the foreground, don't restart it if it exits
  kubectl run -i -t busybox --image=busybox --restart=Never
  
  # Start the nginx pod using the default command, but use custom arguments (arg1 .. argN) for that command
  kubectl run nginx --image=nginx -- <arg1> <arg2> ... <argN>
  
  # Start the nginx pod using a different command and custom arguments
  kubectl run nginx --image=nginx --command -- <cmd> <arg1> ... <argN>
```

{{ header }} Options

```shell
      --allow-missing-template-keys=true
            If true, ignore any errors in templates when a field or map key is missing in the       
            template. Only applies to golang and jsonpath output formats.
      --annotations=[]
            Annotations to apply to the pod.
      --attach=false
            If true, wait for the Pod to start running, and then attach to the Pod as if `kubectl   
            attach ...` were called.  Default false, unless `-i/--stdin` is set, in which case the  
            default is true. With `--restart=Never` the exit code of the container process is       
            returned.
      --command=false
            If true and extra arguments are present, use them as the `command` field in the         
            container, rather than the `args` field which is the default.
      --dry-run='none'
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
      --env=[]
            Environment variables to set in the container.
      --expose=false
            If true, create a ClusterIP service associated with the pod.  Requires `--port`.
      --field-manager='kubectl-run'
            Name of the manager used to track field ownership.
      --image=''
            The image for the container to run.
      --image-pull-policy=''
            The image pull policy for the container.  If left empty, this value will not be         
            specified by the client and defaulted by the server.
  -l, --labels=''
            Comma separated labels to apply to the pod. Will override previous values.
      --leave-stdin-open=false
            If the pod is started in interactive mode or with stdin, leave stdin open after the     
            first attach completes. By default, stdin will be closed after the first attach         
            completes.
  -o, --output=''
            Output format. One of: (json, yaml, name, go-template, go-template-file, template,      
            templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --override-type='merge'
            The method used to override the generated object: json, merge, or strategic.
      --overrides=''
            An inline JSON override for the generated object. If this is non-empty, it is used to   
            override the generated object. Requires that the object supply a valid apiVersion field.
      --pod-running-timeout=1m0s
            The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one    
            pod is running
      --port=''
            The port that this container exposes.
      --privileged=false
            If true, run the container in privileged mode.
  -q, --quiet=false
            If true, suppress prompt messages.
      --restart='Always'
            The restart policy for this Pod.  Legal values [Always, OnFailure, Never].
      --rm=false
            If true, delete the pod after it exits.  Only valid when attaching to the container,    
            e.g. with `--attach` or with `-i/--stdin`.
      --save-config=false
            If true, the configuration of current object will be saved in its annotation.           
            Otherwise, the annotation will be unchanged. This flag is useful when you want to       
            perform kubectl apply on this object in the future.
      --show-managed-fields=false
            If true, keep the managedFields when printing objects in JSON or YAML format.
  -i, --stdin=false
            Keep stdin open on the container in the pod, even if nothing is attached.
      --template=''
            Template string or path to template file to use when -o=go-template,                    
            -o=go-template-file. The template format is golang templates                            
            [http://golang.org/pkg/text/template/#pkg-overview].
  -t, --tty=false
            Allocate a TTY for the container in the pod.
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

