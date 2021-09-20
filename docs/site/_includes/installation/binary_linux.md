```shell
curl -L "https://tuf.werf.io/targets/releases/{{ include.version }}/linux-{{ include.arch }}/bin/werf" -o /tmp/werf
sudo install /tmp/werf /usr/local/bin/werf
```
