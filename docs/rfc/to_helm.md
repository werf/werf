Move `dapp kube deploy` functions into `helm` and improve `helm`.

# Helm install/upgrade/rollback logging

Existed PRs related to helm logging:

* https://github.com/kubernetes/helm/pull/2386
* https://github.com/kubernetes/helm/pull/2342
* https://github.com/kubernetes/helm/pull/3263

## Hooks logs streaming during install/upgrade/rollback helm operation

Helm issue: https://github.com/kubernetes/helm/issues/3481

Helm PR: https://github.com/kubernetes/helm/pull/3479

Almost done, feedback needed.

Helm-client will print logs for each hook with `helm.sh/watch=true` annotation like this:

```
==> Job "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2", Pod "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2-jjpcz", Container "mycontainer-2" <==
mycontainer 2 1
mesg: ttyname failed: Inappropriate ioctl for device

==> Job "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2", Pod "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2-jjpcz", Container "mycontainer-1" <==
mesg: ttyname failed: Inappropriate ioctl for device
mycontainer 1 1
mycontainer 1 2
mycontainer 1 3
mycontainer 1 4
mycontainer 1 5

==> Job "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2", Pod "check-release-hook-b4a465c0-4672-4ecc-9b8c-0e495f497aa2-jjpcz", Container "mycontainer-2" <==
mycontainer 2 2
mycontainer 2 3
mycontainer 2 4
mycontainer 2 5

==> Job "myhook2", Pod "myhook2-zqtfn", Container "mycontainer-1" <==
mesg: ttyname failed: Inappropriate ioctl for device
mycontainer 1 1

==> Job "myhook2", Pod "myhook2-kjqqp", Container "mycontainer-1" <==
mesg: ttyname failed: Inappropriate ioctl for device
mycontainer 1 1

==> Job "myhook2", Pod "myhook2-zqtfn", Container "mycontainer-1" <==
mycontainer 1 2

==> Job "myhook2", Pod "myhook2-kjqqp", Container "mycontainer-1" <==
mycontainer 1 2

==> Job "myhook2", Pod "myhook2-zqtfn", Container "mycontainer-1" <==
mycontainer 1 3

==> Job "myhook2", Pod "myhook2-kjqqp", Container "mycontainer-1" <==
mycontainer 1 3

==> Job "myhook2", Pod "myhook2-zqtfn", Container "mycontainer-1" <==
mycontainer 1 4

==> Job "myhook2", Pod "myhook2-kjqqp", Container "mycontainer-1" <==
mycontainer 1 4

==> Job "myhook2", Pod "myhook2-zqtfn", Container "mycontainer-1" <==
mycontainer 1 5

==> Job "myhook2", Pod "myhook2-kjqqp", Container "mycontainer-1" <==
mycontainer 1 5
NAME:   ex-helm-1-fck
LAST DEPLOYED: Thu Feb 22 18:53:54 2018
NAMESPACE: fck
STATUS: DEPLOYED

RESOURCES:
==> v1/Pod(related)
NAME                       READY  STATUS             RESTARTS  AGE
myserver-7797cc8bfd-5w694  0/1    ContainerCreating  0         0s
myjob-l8cl4                0/1    ContainerCreating  0         0s

==> v1/ConfigMap
NAME         DATA  AGE
myconfigmap  1     0s

==> v1beta1/Deployment
NAME      DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
myserver  1        1        1           0          0s

==> v1/Job
NAME   DESIRED  SUCCESSFUL  AGE
myjob  1        0           0s
```

Logs from parallel containers and pods from hook-job will be shown in realtime with different headers like `tail -f` on multiple files.

# Move dapp/recreate hook annotation to helm

Helm PR: https://github.com/kubernetes/helm/pull/3540

Done, feedback needed.

# Repair helm build and `helm init --wait`

Helm PR: https://github.com/kubernetes/helm/pull/3506

# Helm autorollback

TODO

## Helm rollback experiment

* `helm install` => success
    * Release created
    * `helm list`
        * DEPLOYED state
        * REVISION=1
* `helm upgrade` => failed in post-upgrade hook
    * `helm list`
        * DEPLOYED state
        * REVISION=1
    * Actual resources state in kubernetes is not corresponding to REVISION=1. Part of resources has the new state, part has old state.
* `helm rollback <relese> 1` => success
    * `helm list` shows 2 releases in DEPLOYED state
        * REVISION=1 -- it looks like BUG.
        * REVISION=5 -- it is factual revision, that corresponds to resources state in kubernetes.
* `helm upgrade` => success
    * `helm list` shows 2 releases in DEPLOYED state
        * REVISION=5 -- old revision, not factual anymore, it looks like BUG.
        * REVISION=6 -- new factual revision
* After post-upgrade hook failure:
    * Release is in DEPLOYED state.
    * `helm list -a` shows new revision in PENDING_UPGRADE state.

### Conclusions

* Rollback rolls back resources to previous working state.
* Autorollback is the function of a tiller, because the upgrade will be more atomical that way.
* Without autorollback `helm upgrade` will break your cluster and leave it in incorrect state.

# Helm resource env-variables has not updated after upgrade

TODO: experiment & description here

# 2-way merge problems

Addressed in document https://github.com/thomastaylor312/helm-3-crd#changes-from-helm-2:

> State is no longer stored and releases are now diff'd against the current state in the cluster.

