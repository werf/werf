<div class="details active">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize
jobs:

  converge:
    name: Converge
    if: ${{ contains( github.head_ref, 'review' ) }}
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Define environment url
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_SET_ENV_URL::global.env_url=http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Converge
        uses: werf/actions/converge@master
        with:
          env: review-${{ github.event.number }}
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
```
{% endraw %}

</div>
</div>

<div class="details active">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    if: ${{ contains( github.head_ref, 'review' ) }}
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Dismiss
        uses: werf/actions/dismiss@master
        with:
          env: review-${{ github.event.number }}
          kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
```
{% endraw %}

</div>
</div>