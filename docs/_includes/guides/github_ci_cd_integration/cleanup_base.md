{% raw %}
```yaml
name: Cleanup Docker registry
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Cleanup
        uses: werf/actions/cleanup@master
        with:
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }} }}
```
{% endraw %}