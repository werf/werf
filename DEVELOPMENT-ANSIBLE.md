## DEVELOPMENT ANSIBLE

### Additional python sources

werf repo contains additional callbacks for Ansible in `pkg/build/builder/ansible` directory. These *.py files
are converted into go sources with [esc](https://github.com/mjibson/esc) utility and commited to vcs.

Changes commit process:

1. Install `esc`:

```
go get -u github.com/mjibson/esc
```

2. Go generate:

`pkg/build/builder/ansible` contains go generate line:

```
//go:generate esc -no-compress -ignore static.go -o ansible/static.go -pkg ansible ansible
```

To update `static.go` run from repo root:

```
go generate ./...
```

### Ease development

`werf.py` and `live.py` are copied into tmp directory that mounted into stage container. To ease
development of ansible callbacks, werf can create hardlinks for werf.py and live.py if
`WERF_DEBUG_ANSIBLE_WERF_PY_PATH` or `WERF_DEBUG_ANSIBLE_LIVE_PY_PATH` environment variables are set.

Hardlinks are not possible between different devices or filesystems, so you need to use --tmp-dir flag.
Also note that IDEA based and some other editors can break hardlinks if `safe write` option is enabled.

Example:
```
$ cat werf.yaml

project: app
configVersion: 1
---
image: ~
from: ubuntu:16.04
ansible:
  install:
  - name: add group app
    group:
      name: app
      gid: 7000
  # error
  - name: error
    raw: bash -c 'exit 1'

```

```
$ export WERF_DEBUG_ANSIBLE_WERF_PY_PATH=$REPO_ROOT/pkg/build/builder/ansible/werf.py
$ cd test-project
$ werf build --introspect-error --tmp-dir `pwd`/tmp
```

Now you can edit werf.py in repo tree and run playbook for a stage with command like:

```
root@d49822e52346:/# /.dapp/deps/ansible/2.4.4.0-10/embedded/bin/ansible-playbook /.werf/ansible-workdir/playbook.yml
```

When changes are ready, run `go generate ./...` from $REPO_ROOT and commit static.go and *.py files.
