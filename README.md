<p align="center">
  <img src="https://werf.io/assets/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://github.com/werf/werf/discussions"><img src="https://img.shields.io/static/v1?label=GitHub&message=discussions&color=brightgreen&logo=github" alt="GH Discussions"/></a>
  <a href="https://twitter.com/werf_io"><img src="https://img.shields.io/static/v1?label=Twitter&message=page&color=blue&logo=twitter" alt="Twitter"/></a>
  <a href="https://t.me/werf_io"><img src="https://img.shields.io/static/v1?label=Telegram&message=chat&logo=telegram" alt="Telegram chat"></a><br>
  <a href="https://godoc.org/github.com/werf/werf"><img src="https://godoc.org/github.com/werf/werf?status.svg" alt="GoDoc"></a>
  <a href="https://codeclimate.com/github/werf/werf/test_coverage"><img src="https://api.codeclimate.com/v1/badges/bac6f23d5c366c6324b5/test_coverage" /></a>
  <a href="CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/2503"><img src="https://bestpractices.coreinfrastructure.org/projects/2503/badge"></a>
  <a href="https://artifacthub.io/packages/search?repo=werf"><img src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/werf" alt="Artifact Hub"></a>
</p>

werf is a CNCF Sandbox CLI tool to implement full-cycle CI/CD to Kubernetes easily. werf integrates into your CI system and leverages familiar and reliable technologies, such as Git, Dockerfile, Helm, and Buildah.

What makes werf special:

- **Complete application lifecycle management**: build and publish container images, test, deploy an application to Kubernetes, distribute release artifacts and clean up the container registry.
- **Ease of use**: use Dockerfiles and Helm chart for configuration and let werf handle all the rest.
- **Advanced features**: automatic build caching and content-based tagging, enhanced resource tracking and extra capabilities in Helm, a unique container registry cleanup approach, and more.
- **Gluing common technologies**: Git, Buildah, Helm, Kubernetes, and your CI system of choice.
- **Production-ready**: werf has been used in production since 2017; thousands of projects rely on it to build & deploy various apps.

## Quickstart

The [quickstart guide](https://werf.io/documentation/quickstart.html) shows how to set up the deployment of an example application (a cool voting app in our case) using werf.

## Installation

The [installation guide](https://werf.io/installation.html) helps set up and use werf both locally and in your CI system.

## Documentation

Detailed usage and reference for werf are available in [documentation](https://werf.io/documentation/) in multiple languages.

Developers can get all the necessary knowledge about application delivery in Kubernetes (including basic understanding of K8s primitives) in the [werf guides](https://werf.io/guides.html). They provide ready-to-use examples for popular frameworks, including Node.js (JavaScript), Spring Boot (Java), Django (Python), Rails (Ruby), and Laravel (PHP).

## Community & support

Please feel free to reach developers/maintainers and users via [GitHub Discussions](https://github.com/werf/werf/discussions) for any questions regarding werf.

Your issues are processed carefully if posted to [issues at GitHub](https://github.com/werf/werf/issues).

You're also welcome to:
* follow [@werf_io](https://twitter.com/werf_io) to stay informed about all important news, new articles, etc;
* join our Telegram chat for announcements and ongoing talks: [werf_io](https://t.me/werf_io). _(There is a Russian-speaking Telegram chat [werf_ru](https://t.me/werf_ru) as well.)_

## Contributing

This [contributing guide](https://github.com/werf/werf/blob/main/CONTRIBUTING.md) outlines the process to help get your contribution accepted.

## License

Apache License 2.0, see [LICENSE](LICENSE).

## Featured in

<p>
  <a href="https://console.dev" title="Visit Console - the best tools for developers"><img src="https://console.dev/img/badges/1.0/svg/console-badge-logo-dark-border.svg" alt="Console - Developer Tool of the Week" /></a>
  <a href="https://thenewstack.io/werf-automates-kubernetes-based-gitops-workflows-from-the-command-line/" title="WERF Automates Kubernetes-based GitOps from the Command Line"><img alt="Scheme" src="https://raw.githubusercontent.com/werf/werf/main/docs/images/thenewstack.svg" height="54px"></a>
</p>
