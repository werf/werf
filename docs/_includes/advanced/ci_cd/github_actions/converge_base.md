{% raw %}
```yaml
converge:
  name: Converge
  runs-on: ubuntu-latest
  steps:

    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0

    - name: Converge
      uses: werf/actions/converge@v1.2
      with:
        env: ANY_ENV_NAME
        kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
      env:
        WERF_SET_ENV_URL: "envUrl=ANY_ENV_URL"
```
{% endraw %}
