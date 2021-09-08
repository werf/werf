```shell
curl -LO https://tuf.werf.dev/targets/releases/{{ include.version }}/darwin-{{ include.arch }}/bin/werf
sudo install ./werf /usr/local/bin/werf
```
