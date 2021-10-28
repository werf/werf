
# werf.io site backend

HTTP-server to make some logic for website.

To debug (dlv, 2345/tcp):
```
werf compose up --config werf-debug.yaml --follow --docker-compose-command-options="-d" --docker-compose-options='-f docker-compose-debug.yml'
```
