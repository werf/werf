<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches:
    - master
jobs:

  converge:
    name: Converge
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_SET_ENV_URL::global.env_url=http://${github_repository_id}.kube.DOMAIN

      - name: Converge
        uses: flant/werf-actions/converge@v1
        with:
          env: staging
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    branches:
    - production
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
        uses: flant/werf-actions/converge@v1
        with:
          env: production
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        env:
          WERF_SET_ENV_URL: "global.env_url=https://www.company.org"
```
{% endraw %}

</div>
</div>