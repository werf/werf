```shell
curl -L https://storage.yandexcloud.net/werf/targets/releases/{{ include.version }}/werf-linux-amd64-{{ include.version }} -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```
