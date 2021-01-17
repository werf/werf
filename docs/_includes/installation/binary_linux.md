```shell
curl -L https://dl.bintray.com/flant/werf/{{ include.version }}/werf-linux-amd64-{{ include.version }} -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```
