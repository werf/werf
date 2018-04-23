# Setup osx-based development environment for dapp

## Host system requirements

* $GOPATH environment variable should be set up
* Dapp project should reside in $GOPATH/src/github.com/flant/dapp

### Mounts

Following mounts are used in this configuration:

* `/tmp` -> `/tmp`
* `$GOPATH` -> `/Users/vagrant/go-workspace`

### Docker

Docker in osx is used through tcp. `DOCKER_HOST` env variable is set in `~/.bashrc`.
Tmp is mounted from host system to be shared with docker-server.
To use dapp mount from `build_dir` option `--build-dir=/tmp/dapp_build` required for `dapp dimg build` command.

## How to run

1. Run host-system preparation script:

```
./pre-vagrant-prepare
```

   * Docker daemon will listen on tcp-port
   * Nfs export mount directories will be set up in /etc/exports

2. Run vagrant:

```
vagrant up
```
