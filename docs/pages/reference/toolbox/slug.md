---
title: Slug
sidebar: documentation
permalink: documentation/reference/toolbox/slug.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

In some cases, text from environment variables or parameters can't be used as is because it can contain unacceptable symbols.

To take into account restrictions for docker images names, helm releases names and Kubernetes namespaces werf applies unified slug algorithm when producing these names. This algorithm excludes unacceptable symbols from an arbitrary text and guarantees the uniqueness of the result for each unique input.

There are 3 types of slug built into werf:

1. Helm Release name slug.
2. Kubernetes Namespace slug.
3. Docker tag slug.

There are commands for each type of slug available which apply algorithms for provided input text. You can use these commands upon your needs.

## Basic algorithm

werf checks the text for compliance with slug **requirements**, and if text complies with slug requirements — werf does not modify it. Otherwise, werf performs **transformations** of the text to comply the requirements and add a dash symbol followed by a hash suffix based on the source text. A hash algorithm is a [MurmurHash](https://en.wikipedia.org/wiki/MurmurHash).

The following steps perform, when werf applies transformations of the text in slug:
* Converting UTF-8 latin characters to their ASCII counterpart;
* Replacing some special symbols with dash symbol (`~><+=:;.,[]{}()_&`);
* Removing all non-recognized characters (leaving lowercase alphanumerical characters and dashes);
* Removing starting and ending dashes;
* Reducing multiple dashes sequences to one dash.
* Trimming the length of the data so that result will fit maximum bytes limit.

The transformations are the same for all slugs, because these transformations are restricted enough to be compatible with any of the slug requirements.

### Helm Release name requirements
* it has only alphanumerical ASCII characters, underscores and dashes;
* it contains no more, than 53 bytes.

### Kubernetes Namespace requirements (which are [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements)
* it has only alphanumerical ASCII characters and dashes;
* it contains no more, than 63 bytes.

### Docker Tag requirements
* it must be valid ASCII and may contain lowercase and uppercase ASCII characters, digits, underscores, periods and dashes;
* it contains no more, than 128 bytes.

## Usage

Slug can be applied to arbitrary string with [`werf slugify` command]({{ site.baseurl }}/documentation/cli/toolbox/slugify.html).

Also werf applies slug automatically when used in CI/CD systems such as GitLab CI. See [plugging into CI/CD]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html) for details. The main principles are:
 * apply slug automatically to params that are derived automatically from CI/CD systems environment;
 * do not apply slug automatically to params that are specified manually with `--tag-*`, `--release` or `--namespace`, this way params are only validated to confirm with the requirements.

To apply slug to params specified manually with `--tag-*`, `--release` or `--namespace` user should call [`werf slugify` command]({{ site.baseurl }}/documentation/cli/toolbox/slugify.html) explicitly, for example:

```shell
werf publish --tag-git-branch $(werf slugify --format docker-tag "Features/MyBranch#123") ...
werf deploy --release $(werf slugify --format helm-release "MyProject/1") --namespace $(werf slugify --format kubernetes-namespace "MyProject/1") ...
```
