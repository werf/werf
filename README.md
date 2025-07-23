<p align="center">
  <img src="https://werf.io/assets/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://github.com/werf/werf/discussions"><img src="https://img.shields.io/static/v1?label=GitHub&message=discussions&color=brightgreen&logo=github" alt="GH Discussions"/></a>
  <a href="https://twitter.com/werf_io"><img src="https://img.shields.io/static/v1?label=Twitter&message=page&color=blue&logo=twitter" alt="Twitter"/></a>
  <a href="https://t.me/werf_io"><img src="https://img.shields.io/static/v1?label=Telegram&message=chat&logo=telegram" alt="Telegram chat"></a><br>
  <a href="https://godoc.org/github.com/werf/werf"><img src="https://godoc.org/github.com/werf/werf?status.svg" alt="GoDoc"></a>
  <a href="https://qlty.sh/gh/werf/projects/werf"><img src="https://qlty.sh/gh/werf/projects/werf/coverage.svg" alt="Code Coverage" /></a>
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

## Installation

The [Getting Started guide](https://werf.io/getting_started/) helps set up and use werf both locally and in your CI system.

## Documentation

Detailed usage and reference for werf are available in [documentation](https://werf.io/docs/) in multiple languages.

Developers can get all the necessary knowledge about application delivery in Kubernetes (including basic understanding of K8s primitives) in the [werf guides](https://werf.io/guides.html). They provide ready-to-use examples for popular frameworks, including Node.js (JavaScript), Spring Boot (Java), Django (Python), Rails (Ruby), and Laravel (PHP).

## Community & support

Please feel free to reach developers/maintainers and users via [GitHub Discussions](https://github.com/werf/werf/discussions) for any questions regarding werf. You're also welcome on [Stack Overflow](https://stackoverflow.com/questions/tagged/werf): when you tag a question with `werf`, our team is notified and comes to help you.

Your issues are processed carefully if posted to [issues at GitHub](https://github.com/werf/werf/issues).

For questions that may require a more detailed and prompt discussion, you can use:

* [#werf](https://cloud-native.slack.com/archives/CHY2THYUU) channel in the CNCF’s Slack workspace;
* [werf_io](https://t.me/werf_io) Telegram chat. _(There is a Russian-speaking Telegram chat [werf_ru](https://t.me/werf_ru) as well.)_

Follow [@werf_io](https://x.com/werf_io) to stay informed about all important project's news, new articles, etc.

## Contributing

This [contributing guide](https://github.com/werf/werf/blob/main/CONTRIBUTING.md) outlines the process to help get your contribution accepted.

## License

Apache License 2.0, see [LICENSE](LICENSE).
