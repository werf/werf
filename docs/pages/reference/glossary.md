---
title: Glossary
sidebar: reference
permalink: reference/glossary.html
---

## Dappfile
— is a configuration to build docker images by dapp.

## Dapp name

Dapp name is either:

* a last element of git repository path from `remote.origin.url` git config parameter;
* or directory name, where dappfile reside, in the case when no git repository used or `remote.origin.url` parameter is absent.

Dapp name can be explicitly specified with `--name` basic option of the most dapp commands.
