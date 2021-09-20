```shell
curl -L "https://tuf.werf.io/targets/releases/{{ include.version }}/darwin-{{ include.arch }}/bin/werf" -o /tmp/werf
sudo install /tmp/werf /usr/local/bin/werf
```
