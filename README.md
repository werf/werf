<p align="center">
  <img src="https://raw.githubusercontent.com/deckhouse/delivery-kit/refs/heads/docs/readme/upd/docs/images/dk.svg?sanitize=true" style="max-height:100%;" height="75">
</p>

<p align="center">
  <a href="https://github.com/deckhouse/delivery-kit/discussions"><img src="https://img.shields.io/static/v1?label=GitHub&message=discussions&color=brightgreen&logo=github" alt="GH Discussions"/></a>
  <a href="https://godoc.org/github.com/deckhouse/delivery-kit"><img src="https://godoc.org/github.com/deckhouse/delivery-kit?status.svg" alt="GoDoc"></a>
  <a href="https://qlty.sh/gh/deckhouse/projects/delivery-kit"><img src="https://qlty.sh/gh/deckhouse/projects/delivery-kit/coverage.svg" alt="Code Coverage" /></a>
  <a href="CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
</p>

Deckhouse Delivery Kit is a CLI tool (part of [Deckhouse CLI](https://github.com/deckhouse/deckhouse-cli) as `d8 dk`) that provides a native application delivery workflow for Deckhouse Kubernetes Platform:

- **Complete application lifecycle management**: build and publish container images, deploy an application, distribute release artifacts and clean up the container registry.
- **Ease of use**: use Dockerfiles and Helm chart for configuration and let werf handle all the rest.
- **Advanced features**: automatic build caching and content-based tagging, enhanced resource tracking and extra capabilities in Helm, a unique container registry cleanup approach, and more.
- **Gluing common technologies**: Git, Buildah, Helm, Kubernetes, and your CI system of choice.

It is built on top of the Open Source werf project, adding native Deckhouse ecosystem integration and extended supply chain security features.

## License

Apache License 2.0, see [LICENSE](LICENSE).
