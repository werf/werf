<div class="details active">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches: [master]
jobs:

  converge:
    name: Converge
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Converge
        uses: werf/actions/converge@master
        with:
          env: staging
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        env:
          WERF_SET_ENV_URL: "global.env_url=http://staging-company.kube.DOMAIN"
```
{% endraw %}

</div>
</div>

<div class="details active">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  repository_dispatch:
    types: [production_deployment]
jobs:

  converge:
    name: Converge
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Converge
        uses: werf/actions/converge@master
        with:
          env: production
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        env:
          WERF_SET_ENV_URL: "global.env_url=https://www.company.org"
```
{% endraw %}

</div>
</div>