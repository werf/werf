Убедитесь, что Git версии 2.18.0 или новее и [Docker](https://docs.docker.com/get-docker) установлены.

Выполните в командной строке:
```shell
curl -L "https://tuf.werf.io/targets/releases/{{ include.version }}/darwin-{{ include.arch }}/bin/werf" -o /tmp/werf
sudo install /tmp/werf /usr/local/bin/werf
```
