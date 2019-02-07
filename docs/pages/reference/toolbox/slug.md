---
title: Slug
sidebar: reference
permalink: reference/toolbox/slug.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

In some cases, text from environment variables or parameters can't be used as is because it can contain unacceptable symbols.

To take into account restrictions for docker images names, helm releases names and kubernetes namespaces werf applies unified slug algorithm when producing these names. This algorithm excludes unacceptable symbols from an arbitrary text and guarantees the uniqueness of the result for each unique input.

There are 3 types of slug built into Werf:

1. Helm Release name slug.
2. Kubernetes Namespace slug.
3. Docker tag slug.

There are commands for each type of slug available which apply algorithms for provided input text. You can use these commands upon your needs.

## Basic algorithm

Werf checks the text for compliance with slug **requirements**, and if text complies with slug requirements — werf doesn't modify it. Otherwise, werf performs **transformations** of the text to comply the requirements and add a dash symbol followed by a hash suffix based on the source text. A hash algorithm is a [MurmurHash](https://en.wikipedia.org/wiki/MurmurHash).

The requirements are different for each type of slug.

Helm Release name requirements:
* it has only alphanumerical ASCII characters, underscores and dashes;
* it contains no more, than 53 bytes.

Kubernetes Namespace requirements (which are [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements):
* it has only alphanumerical ASCII characters and dashes;
* it contains no more, than 63 bytes.

Docker Tag requirements:
* it must be valid ASCII and may contain lowercase and uppercase ASCII characters, digits, underscores, periods and dashes;
* it contains no more, than 128 bytes.

The transformations are the same for all slugs, because these transformations are restricted enough to be compatible with any of the slug requirements.

The following steps perform, when werf apply transformations of the text in slug:
* Converting UTF-8 latin characters to their ASCII counterpart;
* Replacing some special symbols with dash symbol (`~><+=:;.,[]{}()_&`);
* Removing all non-recognized characters (leaving lowercase alphanumerical characters and dashes);
* Removing starting and ending dashes;
* Reducing multiple dashes sequences to one dash.
* Trimming the length of the data so that result will fit maximum bytes limit.

## Slugify

{% include /cli/werf_slugify.md %}
