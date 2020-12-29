---
title: Names slug algorithm
sidebar: documentation
permalink: documentation/internals/names_slug_algorithm.html
---

In some cases, environment variables or parameters can't be used as-is since they contain invalid characters.

To meet the requirements for naming docker images, helm releases, and Kubernetes namespaces, werf applies a unified slug algorithm when generating names. This algorithm excludes unacceptable symbols and ensures the uniqueness of the result for each unique input.

There are three types of slugs embedded in werf:

1. Helm Release name slug.
2. Kubernetes Namespace slug.
3. Docker tag slug.

There are commands for applying every type of slug (you can use them depending on your needs). They apply algorithms to the provided input text.

## Basic algorithm

werf checks if an input meets slug **requirements**. If an input complies with them, werf leaves it unaltered. Otherwise, werf **transforms** characters to comply with the requirements while adding a dash symbol followed by a source-based hash suffix at the end. The [MurmurHash](https://en.wikipedia.org/wiki/MurmurHash) hashing algorithm is used.

While transforming the input in the slug, werf performs the following actions:
* Converts UTF-8 Latin characters to their ASCII counterparts;
* Replaces certain special symbols with a dash (`~><+=:;.,[]{}()_&`);
* Removes non-recognized characters (only lowercase alphanumerical characters and dashes remain);
* Removes starting and ending dashes;
* Replaces sequences of dashes with a single dash.
* Trims the length of the string so that result stays within the maximum bytes limit.

The actions are the same for all slugs since they are restrictive enough to satisfy the requirements of any slug.

### The requirements for naming Helm Releases
* only alphanumerical ASCII characters, underscores, and dashes are allowed;
* the length is limited to 53 bytes.

### The requirements for naming Kubernetes Namespaces (a [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements)
* only alphanumerical ASCII characters, and dashes are allowed;
* the length is limited to 63 bytes.

### The requirements for naming Docker Tags
* the tag must be made of valid ASCII and can contain lowercase and uppercase ASCII characters, digits, underscores, periods, and dashes;
* the length is limited to 128 bytes.

## Usage

The slug may be applied to an arbitrary string via the [`werf slugify` command]({{ "documentation/reference/cli/werf_slugify.html" | true_relative_url: page.url }}).

werf automatically applies slug in CI/CD systems such as GitLab CI. See [plugging into CI/CD]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url: page.url }}) for more details. The basic principles are:
 * the slug is auto-applied to parameters that are automatically obtained from the environment of CI/CD systems;
 * the slug isn't auto-applied to parameters that are specified manually via `--release` or `--namespace`; in this case, parameters are only validated to comply with the requirements.

The user should run the [`werf slugify` command]({{ "documentation/reference/cli/werf_slugify.html" | true_relative_url: page.url }}) explicitly to apply slug to parameters specified manually with `--release`, or `--namespace` user should call, for example:

```shell
werf converge --release $(werf slugify --format helm-release "MyProject/1") --namespace $(werf slugify --format kubernetes-namespace "MyProject/1") ...
```
