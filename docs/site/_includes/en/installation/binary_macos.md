Make sure you have Git 2.18.0 or newer and [Docker](https://docs.docker.com/get-docker) installed.

Execute in shell:
```shell
curl -L "https://tuf.werf.io/targets/releases/{{ include.version }}/darwin-{{ include.arch }}/bin/werf" -o /tmp/werf
sudo install /tmp/werf /usr/local/bin/werf
```
