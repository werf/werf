{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Drain node in preparation for maintenance.

 The given node will be marked unschedulable to prevent new pods from arriving. &#39;drain&#39; evicts the pods if the API server supports [https://kubernetes.io/docs/concepts/workloads/pods/disruptions/](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/) . Otherwise, it will use normal DELETE to delete the pods. The &#39;drain&#39; evicts or deletes all pods except mirror pods (which cannot be deleted through the API server).  If there are daemon set-managed pods, drain will not proceed without --ignore-daemonsets, and regardless it will not delete any daemon set-managed pods, because those pods would be immediately replaced by the daemon set controller, which ignores unschedulable markings.  If there are any pods that are neither mirror pods nor managed by a replication controller, replica set, daemon set, stateful set, or job, then drain will not delete any pods unless you use --force.  --force will also allow deletion to proceed if the managing resource of one or more pods is missing.

 &#39;drain&#39; waits for graceful termination. You should not operate on the machine until the command completes.

 When you are ready to put the node back into service, use kubectl uncordon, which will make the node schedulable again.

 [https://kubernetes.io/images/docs/kubectl_drain.svg](https://kubernetes.io/images/docs/kubectl_drain.svg)

{{ header }} Syntax

```shell
werf kubectl drain NODE [options]
```

{{ header }} Examples

```shell
  # Drain node "foo", even if there are pods not managed by a replication controller, replica set, job, daemon set or stateful set on it
  kubectl drain foo --force
  
  # As above, but abort if there are pods not managed by a replication controller, replica set, job, daemon set or stateful set, and use a grace period of 15 minutes
  kubectl drain foo --grace-period=900
```

{{ header }} Options

```shell
      --chunk-size=500
            Return large lists in chunks rather than all at once. Pass 0 to disable. This flag is   
            beta and may change in the future.
      --delete-emptydir-data=false
            Continue even if there are pods using emptyDir (local data that will be deleted when    
            the node is drained).
      --disable-eviction=false
            Force drain to use delete, even if eviction is supported. This will bypass checking     
            PodDisruptionBudgets, use with caution.
      --dry-run='none'
            Must be "none", "server", or "client". If client strategy, only print the object that   
            would be sent, without sending it. If server strategy, submit server-side request       
            without persisting the resource.
      --force=false
            Continue even if there are pods not managed by a ReplicationController, ReplicaSet,     
            Job, DaemonSet or StatefulSet.
      --grace-period=-1
            Period of time in seconds given to each pod to terminate gracefully. If negative, the   
            default value specified in the pod will be used.
      --ignore-daemonsets=false
            Ignore DaemonSet-managed pods.
      --pod-selector=''
            Label selector to filter pods on the node
  -l, --selector=''
            Selector (label query) to filter on
      --skip-wait-for-delete-timeout=0
            If pod DeletionTimestamp older than N seconds, skip waiting for the pod.  Seconds must  
            be greater than 0 to skip.
      --timeout=0s
            The length of time to wait before giving up, zero means infinite
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
            The name of the kubeconfig context to use
      --insecure-skip-tls-verify=false
            If true, the server`s certificate will not be checked for validity. This will make your 
            HTTPS connections insecure
      --kubeconfig=''
            Path to the kubeconfig file to use for CLI requests.
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
      --token=''
            Bearer token for authentication to the API server
      --user=''
            The name of the kubeconfig user to use
      --username=''
            Username for basic authentication to the API server
      --warnings-as-errors=false
            Treat warnings received from the server as errors and exit with a non-zero exit code
```

